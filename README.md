# GoBlog - My website written in Go
GoBlog is a website with blog support written in Go, GoBlog retrieves its content from Notion using it as a CMS, Under the hood GoBlog uses [templ](https://github.com/a-h/templ) and Tailwind CSS to build the pages' HTML which is then managed by the application as a **Webserver** to provide updated content or **Static Site Generator**.

# Usage
You can implement a **Provider** yourself or have a database in Notion to use the **NotionProvider** that provides the data GoBlog expects, follow these steps:
1. Use [this page](https://so-heil.notion.site/965985b012e8476a9e2c23f8aaa8c47f?v=d2dd1883c3cf48babc9a1cbac883928e&pvs=4) as a template for your database
2. Add a [Notion Integration](https://developers.notion.com/docs/create-a-notion-integration) that has access to the database
3. Set environment variables NOTION_API_KEY to the integration secret and NOTION_ARTICLE_DATABASE_ID to the database you created
4. Install the dependencies with `make install-dependencies`
5. Start the web server with `make start`, there is a dev version available with `make dev` that uses [air](https://github.com/cosmtrek/air)

## Architecture - How it's implemented
GoBlog uses a couple of components, these components are represented as Go Packages and Types. This is an overview of the project's architecture represented as a UML Diagram:

![UML Diagram](https://raw.githubusercontent.com/so-heil/goblog/main/business/assets/static/images/diagram.png)

As you can tell, components have their responsibilities and are de-coupled from one another, this gives us the ability to scale, update, and replace the components when needed without technical overhead.

### website
website is the entry point for GoBlog, it compiles to a binary that can be executed in two paths:
- A Go Webserver that serves the website
- A Static Site Generator that builds every page and saves the resulting HTML in the specified path

website is dependent on app for its functionality.

### app
app is a type defined in website package, app has three methods used by website to either build a Static Site or serve a web server:
- updateStore: this method tries to retrieve all pages from **Provider**, build the pages, and update **Store** with this fresh data, updateStore is dependent on another package(pages) that knows how to fetch and build different pages using **Provider** and it's own Page implementation.
- startWebServer: starts the web server serving website pages, this method creates an HTTP server and lets **frontend.Routes()** to register different routes.
- startSSG: updates store once and then starts the static site generation with **frontend.SSG()**

### Frontend
Frontend is what does the actual work of the webserver and SSG is a provider of data stored in **Store**, note that **Store** is updated in website and Frontend does not write to **Store** has two public methods:
- Routes: Registers all the page routes to an HTTP server
- SSG: Properly copies pages from store and Assets to the target SSG path

### Store
Store is any object that can store, load, and delete a page, store is kept updated by app in webserver mode, and the update function uses a concurrent approach to retrieve page data, build and update the store for faster updates

#### Repository
Store is initially implemented by **Repository**, Repository uses a [BadgerDB](https://github.com/dgraph-io/badger) under the hood for data persistance

### Provider
Provider is any object that can provide the articles and their content from its source, along with special page data, it's used to update **Store**.

#### NotionProvider
**NotionProvider** implements **Provider** giving access to a Notion Database containing articles as its source, the article's content is transformed to HTML by this object and provided to update **Store**.
