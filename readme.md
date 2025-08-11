# Content Management System

This is a Content Management System (CMS) built with Go, designed to manage articles, tags, users, and roles.

## Features

- User authentication (JWT-based)
- Role-based access control (RBAC) using bitwise operator for simplifying the logic
- Article and tag management
- basic User profile
- RESTful API endpoints
- Database migrations using golang-migrate
- Docker support for easy deployment

## Project Structure

- `cmd/server/` - Main application entry point and configuration
- `internal/entity/` - Core domain entities (Article, User, Tag, etc.)
- `internal/params/` - Request/response parameter definitions
- `internal/postgresql/` - Database access and SQL queries
- `internal/rest/` - HTTP handlers and middleware
- `internal/service/` - Business logic and services
- `internal/error/` - Custom error types
- `internal/constanta/` - Constants used throughout the project
- `internal/sharevar/` - Shared variables (e.g., user roles)
- `migrations/` - SQL migration scripts
- `docker-compose.yaml` - Docker Compose configuration
- `Makefile` - Common development commands

## Getting Started

### Prerequisites

- go
- Docker & Docker Compose
- PostgreSQL

### Setup

1. Simply run this command, make sure your port 8080 and 5432 is free. The Ports is used to run the server and to see the DB:

   ```bash
   make up
   ```

2. Introducing mocked user

   in this cms have many permission and the cms use bitwise operator to authenticate the user

   | Permission Name               | Value |
   | ----------------------------- | ----- |
   | ReadDraftedAndArchivedArticle | 1     |
   | CreateArticle                 | 2     |
   | DeleteArticle                 | 4     |
   | UpdateStatusArticle           | 8     |

   - first mocked user is **content writer**. It Combines `ReadDraftedAndArchivedArticle` + `CreateArticle`. so the permission is **8**.

   ```json
   {
     "email": "contentwriter@cms.test",
     "password": "aaa"
   }
   ```

   - first mocked user is **editor**. It Combines `ReadDraftedAndArchivedArticle` + `CreateArticle` + `DeleteArticle` + `UpdateStatusArticle`. so the permission is **15**.

   ```json
   {
     "email": "editor@cms.test",
     "password": "aaa"
   }
   ```

3. Accessing the API
   let's explore the app using docs [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

3.1. **Autentikasi**

- Registrasi Pengguna Baru. access the API [here](http://localhost:8080/swagger/index.html#/Auth/post_auth_register)
- Login Pengguna. access the API [here](http://localhost:8080/swagger/index.html#/Auth/post_auth_login)

  3.2. **Profile - Dilindungi JWT**

- Akses Profil Pengguna. access the API [here](http://localhost:8080/swagger/index.html#/Profile/get_profile)

  3.3. **Artikel - Dilindungi JWT (kecuali GET untuk artikel published)**

- Pembuatan Artikel Baru. access the API [here](http://localhost:8080/swagger/index.html#/articles/post_articles). MUST USE account __contentwriter@cms.test__ or **editor@cms.test**
- Pengambilan Daftar Artikel. access the API [here](http://localhost:8080/swagger/index.html#/articles/get_articles)
- Pengambilan Detail Artikel Terbaru. access the API [here](http://localhost:8080/swagger/index.html#/articles/get_articles__articleID_)
- Pembuatan Versi Artikel Baru. access the API [here](http://localhost:8080/swagger/index.html#/articles/post_articles__articleID__versions) atau arickel juga bisa dibuat dengan reference `article_id` dan `article_version_id` access the API [here](http://localhost:8080/swagger/index.html#/articles/post_articles__articleID__versions__articleVersionID_). MUST USE account __contentwriter@cms.test__ or **editor@cms.test**
- Penghapusan Artikel. access the API [here](http://localhost:8080/swagger/index.html#/articles/delete_articles__articleID_). MUST USE account **editor@cms.test**
- Perubahan Status Versi Artikel. access the API [here](http://localhost:8080/swagger/index.html#/articles/put_articles__articleID__versions__articleVersionID__status). MUST USE account **editor@cms.test**
- Pengambilan Daftar Versi Artikel. access the API [here](http://localhost:8080/swagger/index.html#/articles/get_articles__articleID__versions)
- Pengambilan Detail Versi Artikel Tertentu. access the API [here](http://localhost:8080/swagger/index.html#/articles/post_articles__articleID__versions__articleVersionID_)

  3.4. **Tag**

- Pembuatan Tag Baru. access the API [here](http://localhost:8080/swagger/index.html#/Tags/post_tags) . MUST USE account __contentwriter@cms.test__ or **editor@cms.test**
- Pengambilan Daftar Tag. access the API [here](http://localhost:8080/swagger/index.html#/Tags/get_tags)
- Pengambilan Detail Tag Tertentu. access the API [here](http://localhost:8080/swagger/index.html#/Tags/get_tags__name_)
- Logika - Skor Tren Tag (trending_score) is triggered via articles API, and runs every 10 seconds
- Logika - Skor Hubungan Tag Artikel (article_tag_relationship_score) => triggered via articles API

4. shutdown the application

```
make down
```
