# kube-cli

[![Go Report Card](https://goreportcard.com/badge/github.com/ajdnik/kube-cli)](https://goreportcard.com/report/github.com/ajdnik/kube-cli)
[![License MIT](https://img.shields.io/badge/license-MIT-lightgrey.svg)](https://github.com/ajdnik/kube-cli/blob/master/LICENSE)

The tool automates deployment of projects to Kubernetes using a simple command `kube-cli deploy`. It uses Google Cloud Build service to build Docker images which are then automatically deployed to a chosen Kubernetes cluster running on Google Cloud.

**This tools is in pre-alpha stage. It's not meant to be used in a production environment.**

## Install

**MacOS:**


Paste the following command into a macOS terminal prompt. The script explains what it will do and then pauses before it does it.

```
/usr/bin/ruby -e "$(/usr/bin/curl -fsSL https://raw.githubusercontent.com/ajdnik/kube-cli/master/install/install-macos.rb)"
```

## Running kube-cli

In order to authenticate with Google Cloud the tool uses [Application Default Credentials](https://developers.google.com/identity/protocols/application-default-credentials). The credentials need to be set using the `GOOGLE_APPLICATION_CREDENTIALS` environment variable. For more information, see [Providing credentials to your application.](https://cloud.google.com/docs/authentication/production#providing_credentials_to_your_application)

Make sure the service account used to authenticate the tool has the following roles: *Cloud Build Service Account*, *Kubernetes Engine Admin* and *Storage Admin*.

## Deploying your application

After installing the tool and ensuring the correct `GOOGLE_APPLICATION_CREDENTIALS` environment variable is set you can start using the tool to deploy projects to a Kubernetes cluster. The first step is to configure the project, you can do this by going into the root of your project and running `kube-cli init`. The *init* command will generate a *kubecli.yaml* file in the project root which will serve as a project config. The second step will be to run the `kube-cli deploy` command which will package the project and upload it to Google Cloud Build to build a Docker image, afterwards it will deploy the image to a chosen Kubernetes deployment. Depending on how you've setup the Dockerfile you might have to compile/transpile the binaries or execute some additional steps before running the `kube-cli deploy` command.

**Rollback deployment:**

If you've made a mistake you can always call `kube-cli rollback` which will revert the deployment to it's previous state.

**.kubecliignore file:**

The second step when deploying the project is to archive the entire project directory and upload it to Google Cloud to build the Docker image. Depending on your project the Docker build might not require all of the files in the project folder. In order to control which files/folders get uploaded to Google Cloud you can blacklist files and folders by adding the into the *.kubecliignore* file in the project root.

