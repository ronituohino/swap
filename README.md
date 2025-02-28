# Software Architecture Project - search engine

Group project for the course [Software Architecture Project](https://studies.helsinki.fi/courses/course-implementation/hy-opt-cur-2425-f0bc7662-8185-4d45-a0e1-60e250819047/CSM14103).

A simple search engine for the internet. Built for scalability using Kubernetes. All services are containerized. Uses many existing tools for fast development.

[Roni Tuohino](https://github.com/ronituohino)  
[Perttu Kangas](https://github.com/DeeCaaD)

## Documentation

- [Course Report and Architecture](./docs/report.md)
- [Deployment](./docs/deployment.md)

## Development

- [Kubernetes](./k8s/README.md)

### Running Services Individually

Running services individually requires defining environment variables. Also, some services like the Crawler or Indexer depend on other services such as PostgreSQL and RabbitMQ. Therefore, it is recommended to use local Kubernetes development environment which starts up everything locally.

- [API](./api/README.md)
- [Indexer](./indexer/README.md)
- [Client](./client/README.md)
- [Crawler](./webcrawler/README.md)
