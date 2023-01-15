<a name="readme-top"></a>

# audiotheker 🎶

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li><a href="#about-the-project">About The Project</a></li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
   <li><a href="#license">License</a></li>
  </ol>
</details>


<!-- ABOUT THE PROJECT -->
## About The Project

`audiotheker` allows downloading all episodes of a program/collection or an individual episode in the [ARD Audiothek](https://www.ardaudiothek.de/). It queries the official [GraphQL API](https://api.ardaudiothek.de/docs/#/GraphQL) to gather the download URLs.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

`Go` is required to build the binary **OR** `Docker` can be used to build an image and run a container.

* Go 1.16 or newer

**OR**
* Docker 

### Installation

#### Building

1. Clone the repo
   ```sh
   $ git clone git@github.com:fbngrmr/audiotheker.git
   ```
2. Change directory
   ```sh
   $ cd audiotheker
   ```
3. Build 
   ```sh
   $ make bin
   ```

#### Docker

1. Clone the repo
   ```sh
   $ git clone git@github.com:fbngrmr/audiotheker.git
   ```
2. Change directory
   ```sh
   $ cd audiotheker
   ```
3. Build `Docker` image from `Dockerfile`
   ```sh
   $ docker build -f Dockerfile.amd64 -t audiotheker:0.2.0 .
   ```

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- USAGE EXAMPLES -->
## Usage

Copy the URL to a program, collection, or an individual episode from your browser and provide the URL and a target directory to the binary or a Docker container.

### Built binary

```sh
$ ./build/audiotheker download \
   "https://www.ardaudiothek.de/sendung/j-r-r-tolkien-der-herr-der-ringe-fantasy-hoerspiel-klassiker/12197351/" \
   PATH/TO/YOUR/DOWNLOADS
```

### Docker
```sh
$ docker run \
   --rm \
   --user $(id -u):$(id -g) \
   -v PATH/TO/YOUR/DOWNLOADS:/download \
   audiotheker:0.1.0 download \
   "https://www.ardaudiothek.de/sendung/j-r-r-tolkien-der-herr-der-ringe-fantasy-hoerspiel-klassiker/12197351/" \
   /download
```

<p align="right">(<a href="#top">back to top</a>)</p>

<!-- License -->
## License
`audiotheker` is distributed under Apache-2.0. See LICENSE.

<p align="right">(<a href="#top">back to top</a>)</p>
