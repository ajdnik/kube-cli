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
