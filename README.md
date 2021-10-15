# SmartIDE - The Cloud Native IDE

The Cloud Native designed for boosting development efficiency.

## Release

- v0.1 

## How to Build SmartIDE

### Docs

SmartIDE use [Hugo](https://gohugo.io/) for it's documentation. Theme [Docsy](https://www.docsy.dev) is used for the theme and it's configured as a git submodule, you need to run the following command to recursively clone the submodules in order to work on the project.

```shell
## Clone the main repo
git clone https://github.com/SmartIDE/SmartIDE.git
cd SmartIDE
## Clone the submodules and all its dependencies
git submodule update --init --recursive
## If you have SmartIDE installed, run the following
smartide start
```

Use the following command to start a Hugo Development server to work on the documentation.

> **Note:** 
> - Make sure you have [Hugo](https://gohugo.io/) installed first.
> - If you are running SmartIDE, use the build-in terminal to run the following commands

```shell
npm install
cd docs
hugo server --bind 0.0.0.0
```

Open http://localhost:1313

Now you can simply edit your markdown file and view the site updated on the fly. 

## Copyright 

&copy;[leansoftX.com](https://leansoftx.com) 2021

