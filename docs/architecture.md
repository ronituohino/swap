# Architecture

This document describes the architecture of the search engine.  
The design follows the Microservices pattern language, with inherent traits of scalability and modularism.  

The search engine has three major parts: Crawling, Indexing, and Serving.  
Some parts of the application are in a Kubernetes cluster, while others are run on separate devices.  

## Crawling

The crawler is implemented in [../webcrawler](../webcrawler/).  
It uses the [scrapy](https://scrapy.org/) framework to crawl the internet and scrape keywords from the pages.

_Workflow:_

- The crawler connects and authenticates to RabbitMQ
- Then it starts crawling on some predefined webpages
- It finds links in websites, and follow them to new websites
- Whenever it encounters a new web page:
  - Check if we can crawl this page (is https, robots.txt rules, language="en")
  - If yes, scan the page for words
  - Then [lemmatize](https://en.wikipedia.org/wiki/Lemmatization) the words using a lookup from [this dictionary](https://github.com/michmech/lemmatization-lists/blob/master/lemmatization-en.txt)
  - Assemble a list of most common words in this page, cut off words that are rare
  - Send a message to a RabbitMQ topic ("text")

The format of the message is as follows:

```
{ "example.com": ["domain", "example", ...] }
```

## Indexing

Indexing contains a few parts: RabbitMQ, Indexer, Database.

### RabbitMQ

We use [RabbitMQ](https://www.rabbitmq.com/) as a message broker between the crawlers, and indexers. This allows us to detach the crawlers and indexers from each other architectually. 

_Architectural Effects:_

++ Improves scalability, because we can have multiple crawlers, and multiple indexers, which do not need to know about eachother. This also acts as a load balancer for the indexers, since each message should be handled by only one indexer.  
++ Improves fault-tolerance, because any of the crawlers or indexers can crash at any point, and the system will still function. Also, the nodes in RabbitMQ itself can crash, which might result in some losses in messages, but the system should be able to recover after some time.  
++ Improves extensibility, because we can add new types of crawlers and indexers (e.g. image search) by just creating a new topic in RabbitMQ.  
-- The costs of the system increase, because we need to have more processes constantly running.  
-- The system has more complexity overall. First we wanted to use Kafka, because its more performant according to some sources, but it looks a lot more complex than RabbitMQ. We decided to opt for this one to keep the complexity under control for this project.

### Indexer

The indexer is implemented in [../indexer](../indexer/).  
It uses a high-performance language to create a reverse-index from the messages sent by crawlers. 

For example, the reverse-index for the `example.com` message would look like this:

```
{ "example.com": ["domain", "example", ...] }

└── (gets transformed into) ──┐

{ "domain": ["example.com"], "example": ["example.com"] }
```

_Workflow:_

- The indexer connects and authenticates to Postgres
- The indexer connects and authenticates to RabbitMQ
- Then it subscribes to a RabbitMQ topic ("text")
- Whenever a message is received from RabbitMQ
  - Add data to local reverse-index
  - Start a timer if not active already (e.g. 60s)
  - If new messages arrive, add them to the local index
  - When the timer expries, flush the local index to the database

_Architectural Effects:_

++ Improves the speed of the search engine drastically due to reverse-index, because we don't have to look up every domain if it has the keywords. Now we can instead look up which domains have the keyword.  
++ Reduces load on the database by reducing the amount of requests, and instead sending a lot of data with each request.
-- The system has more complexity overall.

### Database

We use [Postgres](https://www.postgresql.org/) to store our search index.  
The database would have 3 tables:
- `websites`
- `keywords`
- `references`

The database should be optimized using [indexes](https://www.postgresql.org/docs/current/indexes.html).

_Architectural Effects:_

++ Improves accuracy, because it's an ACID database and it uses transactions to save crawl information.  
-- Reduces scalability, because it's a relational database and it doesn't scale horizontally.

## Serving

The search engine is accessible through an API, and we might implement a static client application for it.

### API

The api is a simple REST API, with a single GET endpoint: `/search`.  
The endpoint takes the search query in a query parameter `q` with URI normalization applied to it.

For example, when searching for "cat pictures", the query would be like this:

```
/serch?q=cat+pictures
```

It queries the Postgres database for search results, and returns them as a JSON response like this:

```
{
  "results": [
    { "url": "https://www.catoftheday.com/", ... },
    { "url": "https://www.reddit.com/r/catpictures/", ...},
    ...
  ],
  "query_time": "0.0000001s",
  ...
}

```

### Client app (optional)

If we have time, we could implement a very simple client app to use the API. It could be served from the root of the API.

## Kubernetes

All services from Indexing (RabbitMQ, Indexer, Postgres), and the API, are in a Kubernetes cluster. The definitions for these services are in [../k8s](../k8s/).  
This makes it easy for us to deploy and manage the application and complexity.  

_Architectural Effects:_

++ Increases scalability, because Kubernetes will automatically scale the services up/down according to demand.  
++ Increases maintainability, because there are lots of tools for Kubernetes to manage services.  
-- Increases complexity, because K8s is not an easy system to understand fully. 