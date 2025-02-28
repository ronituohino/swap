# Architecture Description and Course Report

This document describes the architecture of the search engine and serves as a course report.

## Introduction

We are building a general search engine that indexes pages from the internet. The goal is that our search engine returns reasonable results to websites that contain somewhat relevant information to search terms. Our aim is to study the mechanisms used in simple search tasks and study architectural patterns while building one. We wish to index only a fraction of the internet but make an engine that has the potential to scale to index the entire internet.

The design of the search engine follows the Microservices pattern language, with inherent traits of scalability and modularity. We are heavily utilizing [Docker](https://www.docker.com/) and [Kubernetes](https://kubernetes.io/) to containerize our services and orchestrate the application. This allows us to develop each service independently from each other and use whatever technologies we see fit for each task. Furthermore, it helps us divide development tasks because there are clear boundaries between each service.

The experimentation setup section describes the roles of each service of the search engine in detail and the reasoning behind architectural choices.  
In the end, we will also list observations, conclusions, and lessons learned from this project.

## Experimentation Setup

Most of our services are wrapped in a Docker container, which is essentialy a lightweight [virtualization](https://en.wikipedia.org/wiki/Virtualization) technique. These containers are managed by Kubernetes. It groups containers into logical units (Pods) and provides features like automatic deployment, scaling, load balancing, self-healing, automated rollouts/rollbacks, and storage orchestration. This makes it easy for us to manage the application and complexity. However, this comes with some drawbacks. Kubernetes has lots of new concepts (Pods, Deployments, Services, etc...), which have a steep learning curve in order to fully utilize the features that Kubernetes offers. Furthermore, Kubernetes is meant for enterprise-scale applications, which could be a little overkill for this project.

Our services that are in wrapped in Docker containers are deployed in a Kubernetes cluster. A cluster is essentially a collection of physical hardware, that Kubernetes can allocate services onto. The definitions for our services, and Kubernetes cluster are in [../k8s](../k8s/). 

_Architectural Effects:_

++ Increases scalability because Kubernetes can be configured to automatically scale services up/down according to demand.  
++ Increases maintainability because Kubernetes offers lots of tooling to manage services. This is particularly good because some of the standard tooling is very common in the industry.  
-- Increases complexity because Kubernetes has a steep learning curve.

The search engine can be divided into two major parts: Indexing and Serving, which roughly translate to the backend and frontend of the application.

![architecture](./architecture.png)

### Indexing

The first major part of our application is Indexing, which refers to services that are needed to form our search index of the internet.  
Indexing contains 5 services: Crawler, RabbitMQ, Indexer, Database, and IDF.

#### Crawler

The crawler is an application (or "bot") that scans web pages for relevant terms within the page. For example, the relevant terms for a page about making mocha cake would probably be "mocha", "cake", "cooking", "baking". The crawler uses many heuristics to count the `relevance score` for each word. These heuristics include placing a higher relevance score to words that are at the beginning of the web page, or that are wrapped in important [semanting elements](https://www.w3schools.com/html/html5_semantic_elements.asp) such as headers (h1, h2, ...) or the title tag for the web page. After scanning a page for the most relevant terms, the crawler looks up links to other pages, which it then continues to scan. 

The crawler is implemented in [../webcrawler](../webcrawler/), and it uses the [scrapy](https://scrapy.org/) framework.

_Workflow:_

- The crawler connects and authenticates to RabbitMQ
- Then it starts crawling on some predefined web pages
- It finds links in websites and follows them to new websites
- Whenever it encounters a new web page:
  - Check if we can crawl this page (is https, robots.txt rules, language="en")
  - If yes, scan the page for words
  - Preprocess words
    - Transform all words into lowercase
    - Remove words with a single letter, like "i", "a"
    - [Lemmatize](https://en.wikipedia.org/wiki/Lemmatization) words using a lookup from [this dictionary](https://github.com/michmech/lemmatization-lists/blob/master/lemmatization-en.txt)
    - Remove words that do not provide value, like ["and", "then", "is"](https://en.wikipedia.org/wiki/Most_common_words_in_English)
    - [Among other things ...](../webcrawler/src/spiders/infinite_spider.py)
  - Assemble a list of relevant words on this page
  - Send a message to RabbitMQ

The format of the message is as follows:

```json
{
  "url": "https://preppykitchen.com/mocha-cake/",
  "title": "Mocha Cake - Preppy Kitchen",
  "keywords": {
    "mocha": {
      "relevance": 70,
      "term_frequency": 0.01
    },
    "cake": {
      ...
    },
    ...
  }
}
```

#### RabbitMQ

We use [RabbitMQ](https://www.rabbitmq.com/) as a message broker between the Crawlers and Indexers. RabbitMQ offers many ways to implement messaging, but we are specifically using [Worker Queues](https://www.rabbitmq.com/tutorials/tutorial-two-go). Essentially the RabbitMQ brokers take messages in from the producers (web crawlers), and send the messages for processing to consumers (indexers). In the Worker Queue approach, each message is sent to one consumer, and has to be acknowledged that it has been processed correctly. This approach is akin to load balancing, but better since the messages are not lost if something fails. The best part is that this allows us to detach the crawlers and indexers from each other architecturally, which is at the heart of event-driven architecture. We feel this approach is really good and should be adapted wherever possible.

_Architectural Effects:_

++ Improves scalability because we can have multiple crawlers and multiple indexers, which do not need to know about each other.   
++ Improves fault tolerance because any of the crawlers or indexers can crash at any point, and the system will still function. Also, the nodes in RabbitMQ itself can crash, which might result in some losses in messages, but the system should be able to recover after some time.  
++ Improves extensibility because we can add new types of crawlers and indexers (e.g. image search) by just creating a new topic in RabbitMQ.  
-- The costs of the system increase because we need to have more processes constantly running.  
-- The system has more complexity overall. First, we wanted to use [Kafka](https://aws.amazon.com/compare/the-difference-between-rabbitmq-and-kafka/) because it's more performant according to some sources, but it looks a lot more complex than RabbitMQ. We decided to opt for this one to keep the complexity under control for this project.

#### Indexer

The indexer is an application that takes crawl results from Crawlers and processes them into a suitable format (reverse-index) for the database. Instead of storing which words are found on a certain website, we want to store which websites contain a certain word. This is the reverse-index method, which speeds up search times drastically.

In addition, whe indexer implements a buffering feature. Instead of processing a single message at a time, the Indexer takes in a variable amount of messages (100 in our case) and processes them in bulk. This reduces network and database load significantly.

The indexer is implemented in [../indexer](../indexer/).

For example, the reverse-index for the `mocha cake` message would look like this:

```json
{
  "mocha": ["https://preppykitchen.com/mocha-cake/"],
  "cake": ["https://preppykitchen.com/mocha-cake/"]
}
```

_Workflow:_

- The indexer connects and authenticates to Postgres
- The indexer connects and authenticates to RabbitMQ
- Whenever a message is received from RabbitMQ
  - Wait 30s or until message buffer size is reached
  - Convert the buffer to database compliant reverse-index
  - Flush the converted buffer to the database
  - Send acknowledgment of the whole buffer to RabbitMQ

_Architectural Effects:_

++ Improves search speed due to reverse-index because we don't have to look up every domain if it has the keywords. Now we can instead look up which domains have the keyword.  
++ Reduces load on the database by reducing the number of requests and instead sending a lot of data with each request.  
-- Increases system complexity, because one more service to manage in the Kubernetes cluster. In theory the current implementation could work without an Indexer service, but it has lots of benefits and in the future this service would become mandatory as the processing logic grows.

#### Database

We use [Postgres](https://www.postgresql.org/) to store our search index.

The database has 3 tables:

- `websites`
- `keywords`
- `relations`

The websites table stores information about a specific website (url, title). The keywords table stores every english word found during crawling. The relations map the keywords to certain websites, and store information about the relevance of the word on that site.

The database is optimized using [indexes](https://www.postgresql.org/docs/current/indexes.html) to speed up search queries.  

The database schema is found in [../indexer/../db.go](../indexer/internal/db/db.go).

_Architectural Effects:_

++ Simplest and cheapest database solution, because we don't have to worry about data replication. Also there are plenty of free Postgres service providers.  
-- Reduces scalability because it's a relational database and it doesn't scale horizontally. We could use [CockroachDB](https://www.cockroachlabs.com/) or [Cassandra](https://cassandra.apache.org/_/index.html) to make it distributed. However, these are not familiar technologies for us, and we want to get a functional service up as quickly as possible. Therefore in the current state of this project, the main bottleneck is the database.

#### IDF

We use [TF-IDF](https://en.wikipedia.org/wiki/Tf%E2%80%93idf) along with the relevance of words when scoring the search results. This method has 2 calculations, which are multiplied together: `TF` (term frequency), and `IDF` (inverse document frequency). 

TF is calculated by the crawler. The formula for TF is quite simple: `(occurrences of a certain word on a website) / (the total amount of words on a website)`. 

IDF is calculated as a separate cron job. The formula for IDF is more complicated: `log((total number of websites in the index) / (the number of websites this word appears on))` To calculate this, we need to go over the entire database, which is why this is a separate service Preferably IDF calculations would run when the database has low load. In addition, the calculations are done in batch sizes of `1,000,000` to reduce load and the size of a single commit.

_Architectural Effects:_

++ Reduces load on the database due to not having to recalculate IDF for each relation every time new data is inserted  
++ Increases the quality of search results, because rare words get more emphasis in search queries  
-- Further increases system complexity, because a new service in the Kubernetes cluster to manage  
-- Requires a lot of CPU power during calculations

### Serving

The second major part of our application is Serving, which refers to services that are needed to make our search index accessible to people.  
Serving has 2 services: an API and a Client.

#### API

The API serves as the only public entry point into the application. It handles incoming HTTP requests and queries the database for results. In addition it calulates the final score for search results during the query: `TF * IDF * relevance score`

The API is implemented as a simple REST API, with one GET endpoint: `/search`.  
The endpoint takes the search terms in a search parameter `q` with URI normalization applied to it.

For example, when searching for "cat pictures", the request would be like this:

```
/search?q=cat+pictures
```

The API then queries the Postgres database for search results and returns them as a JSON response like this:

```json
{
  "results": [
    {
      "url": "https://www.catoftheday.com/",
      "title": "Cat of the day",
      "score": 0.456542,
      "keywords": [
        "cat",
        ...
      ]
    },
    { "url": "https://www.reddit.com/r/catpictures/", ...},
    ...
  ],
  "query_time": "0.031s",
  "total_hits": 20,
  ...
}
```

#### Client

The Client is a simple static web application that requests our API for search results and presents them nicely. It is implemented using the [Astro](https://astro.build/) framework, and it is hosted on [GitHub Pages](https://pages.github.com/).

It is available at https://ronituohino.github.io/swap/.

The crawler `USER_AGENT` field points to this website to inform what our crawling traffic is about. It is a good manner to identify your crawlers for website owners. Before we formed the index, our site was used as a notice for the websites we were crawling.

## Observations, Lessons learned and Conclusions

### Infrastructure

Setting up the infrastructure for this kind of project requires some work, but proper preplanning helped a lot to map which services handle what tasks. Because of the preplanning, we were able to divide tasks and develop them individually. This flexibility in infrastructure allows us to use tools of our choice, which can be seen in the variety of languages used in the project: Go, Python, TypeScript, Shell and PLpgSQL.

The infrastructure we chose ended up being quite performant in terms of scalability. The main bottleneck currently is the non-distributed database.

### Tools

Searching up tools can speed up the development. For example, first we were going to do our own web crawler implementation which would have been quite slow in its simplest implementation. We ended up using a library for it. If there is a maintained tool for some task, it should at least be considered instead of creating your own implementation.

Using tools like Copilot can speed up the development if used correctly. One must understand the changes it suggests before pressing "Accept". They can also help developers by pushing them in the right direction in solving problems.

### Data management

Data management ended up being the most difficult part of this project. We created our index by scraping the internet for `~4hr` using only **one** web crawler and indexer with the following results:

| Table name | Total count | Total size | Table size | Index size |
| ---------- | ----------- | ---------- | ---------- | ---------- |
| websites   | ~140k       | 44 MB      | 26 MB      | 19 MB      |
| keywords   | ~190k       | 13 MB      | 9.5 MB     | 4.2 MB     |
| relations  | ~17.3m      | 2015 MB    | 1125 MB    | 890 MB     |

Surprisingly, a lot of data fits into the index while maintaining reasonable disk usage.

Our next problem was that queries took a lot of time. When searching `Los Angeles` the search took `~8.83s`. We ended up adding database indexes for the relations table on the `website_id` and `keyword_id` columns, which resulted in the same query taking `~2.83s`, and increasing the relations index size by `236MB`.

| Table name | Total count | Total size | Table size | Index size |
| ---------- | ----------- | ---------- | ---------- | ---------- |
| websites   | ~140k       | 44 MB      | 26 MB      | 19 MB      |
| keywords   | ~190k       | 13 MB      | 9.5 MB     | 4.2 MB     |
| relations  | ~17.3m      | 2251 MB    | 1125 MB    | 1126 MB    |

The next problem we faced was IDF calculations. The first version we had would've taken an unreasonable amount of time (days) for the above table sizes. To improve this we removed the IDF batching to process the entire IDF calculation in one go. In addition, we also increased the `shared_buffers`, `work_mem` and `max_wal_size` in PostgreSQL drastically. This resulted in processing IDF updates for all `~17.3m` relations in `~18min`. We tried to re-add batching with a size of `100,000` which resulted in `~55min` processing time. We ended our experiments there. Batch processing would offer benefits such as smaller transactions, increased system responsiveness, and fewer locks in Postgres.

### Buffers

Processing data in bulks is usually better if there is a lot of data moving around. This reduces network and service load. This can be seen in our scenario quite well:

In the first version of the Indexer, when it received a message from RabbitMQ, it instantly inserted the message into the database. When we started 3 indexers, 3 web crawlers, and 3 RabbitMQ instances in parallel we got Postgres constantly throwing errors about deadlocks.

To reduce this we added a message buffer and flush timer. Now it processes `100` messages or waits for `30s` before inserting them into the database in bulk. The messages are acknowledged to RabbitMQ by the indexer after all of them are inserted into the database. Therefore either all of them succeed or all fail. Postgres threw less errors in this implementation, however they did not go away entirely.
