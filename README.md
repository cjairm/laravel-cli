laravel-cli command
=============

* [![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Description
-----------

`laravel-cli` helps you to create a dockerized [Laravel](https://laravel.com/docs/11.x) (version ^11 ONLY) project with a single command usage

Features
--------

* **Creates Basic Docker templating**: It creates and downloads needed files. 
* **Replaces some variables for custom ones**: As docker file is created, we are able to know which vars the .env documents should use, so we update them.

Table of Contents
-----------------

* [Installation](#installation)
* [Usage](#usage)

### Installation

To install the project, follow these steps (assuming you already installed `Docker`):

Recommended to create new folder directory (if you want the command as global)

```bash
mkdir ~/my-custom/path && cd ~/my-custom/path
```

Clone the repository:

```bash
git clone https://github.com/cjairm/laravel-cli.git

################################################################
### If you want to make it a global command...
echo 'alias laravel-cli="~/my-custom/path"' >> ~/.zshrc
source ~/.zshrc
################################################################
```

### Usage
From here as simple as

```bash
laravel-cli create docker --dir /path/to/my/new-project
```

| Flag name         | Default value       | Description          |
| ----------------- | ------------------- | -------------------- |
| `--appName`, `-n` | Parent folder name  | The name of your app |
| `--appPort`, `-p` | 8000                | The posrt of you app |

To put up your service you can do

```bash
docker-composer up # use --build if it's the first time running it
```

You have available commands for npm, composer and artisan integrated in your app

Composer:
```bash
docker-compose run --rm composer [command]

# Example. docker-compose run --rm composer create-project laravel/laravel:^11.0 .
```

Artisan:
```bash
docker-compose run --rm artisan [command]

# Example. docker-compose run --rm artisan migrate
```

NPM:
```bash
docker-compose run --rm npm [command]

# Example. docker-compose run --rm npm update --no-save
```

Enjoy! :smiley:
