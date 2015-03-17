# Populator

is a tool that downloads Docker image repositories and builds them, based on simple configuration file. You use it as a faster alternative to get images to your local docker, if your image registry is slow (i.e. if you are using Docker Hub).

You can also use this if you don't want to upload your images to Docker Hub and you don't have a private regstry yet.

You can describe your repositories and images in yaml as

```
<repo_url>:
  [ localDir: /home/tomk/my_image_repo ]
  <path_in_repo>: t0mk/image1
  <another_path_in_repo>: t0mk/image2
```

If you supply the `localDir`, populator will clone the repo to that directory. If you don't supply it, populator will get the repo to `$HOME/<random_number>_repo_name`. If the `localDir` already exists, populator will try to `git pull` there.

## Practical example

You just spawned a coreos and you want to have there the tutum mysql 5.5, 5.6 and php images. You want to have the php image repo in `~/tutum-php`. Create example.yml as:

```
'https://github.com/tutumcloud/tutum-docker-php':
  localDir: ~/tutum-php
  ./ : whatever/php

'https://github.com/tutumcloud/tutum-docker-mysql':
  5.5/ : whatevs/mysql5.5
  5.6/ : t0mk/mysql5.6
```

Note that you need to put the repo URLs to quotes, otherwise yaml parser is confused by the colons.

## Usage

```
$ ./populator -config example.yml
```

See `./populator --help` for more. You can only download git repos, or only build images, or you can just build images mathcing some substring. 

For instance, with the `example.yml` above, you can choose to only build mysql 5.5 by running:

```
$ ./populator -config example.yml -onlybuild -only 5.5
```

If you are cloning from private repos and you are asked for username and password too much, run populator with `-credcache` flag.
