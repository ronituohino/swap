# Deployment

This document has details about the live version of the application. It's not
running on Kubernetes, because cloud providers charge quite a bit for a
Kubernetes cluster anywhere. Instead we've deployed the parts responsible for
Serving the application separately.

## Database

The live database is deployed on [Aiven](https://aiven.io/), because they offer
a free Postgres instance with up to 5GB of storage. To get the database out of
the Kubernetes cluster, we used `pg_dump` and `kubectl cp`.

In Aiven, we created a new user for the application, and we had to give
permissions with

```
GRANT SELECT ON keywords, relations, websites TO db_user;
```

In addition, to speed up the query time, we had to manually create some indexes
(these are done automatically by the indexer in the Kubernetes cluster)

```
CREATE INDEX IF NOT EXISTS idx_relations_keyword_id ON relations (keyword_id);
CREATE INDEX IF NOT EXISTS idx_relations_website_id ON relations (website_id);
```

Furthermore, Aiven forces `SSL Mode = require`, so we had to manually export the
CA cert, and configure it for the API process.

## API

The live API process is deployed on
[Cloud Run](https://cloud.google.com/run?hl=en).

We built a new Docker image with the following naming schema to push it to the
[Artifact Registry](https://cloud.google.com/artifact-registry/docs)

```
<artifact registry repository region>-docker.pkg.dev/<google cloud project id>/<artifact registry repository name>/<image name>:<tag>
```

The CA certificate is stored in
[Secrets Manager](https://cloud.google.com/security/products/secret-manager?hl=en),
and it is mounted to the Docker container in Cloud Run.

## Client

The live client is hosted on [GitHub Pages](https://pages.github.com/), and it
simply queries the API process in Google Cloud. The deployment process is in
[GitHub Actions]() in the project repository.

To deploy:

```
Go to repo -> Actions -> Deploy to GitHub Pages -> Run workflow -> Branch: main + Run workflow
```
