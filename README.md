<p align="center">
  <a href="" rel="noopener">
 <img width=200px height=200px src="https://storage.googleapis.com/lg-static-prod/tbl.png" alt="Lake Games Logo"></a>
</p>

<h3 align="center">Lake Games Web App</h3>

<div align="center">

[![Status](https://img.shields.io/badge/status-active-success.svg)]()
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](/LICENSE)


</div>

---

<p align="center"> Lake Games web application built with ReactJS and Django and deployed to Kubernetes using Helm
    <br> 
</p>

## 📝 Table of Contents

- [About](#about)
- [Getting Started](#getting_started)
- [Script Usage](#usage)
- [Built Using](#built_using)
- [CI/CD](#cicd)
- [Authors](#authors)

## 🧐 About <a name = "about"></a>

Lake Games is a website built to promote water sports like wakeboarding and wakesurfing in a certain community. Originated in Sister Lakes, Michigan

## 🏁 Getting Started <a name = "getting_started"></a>

Set environment variables

Backend - get secret variables from admin
```
source ./envs/dev.sh
```

Run frontend server
```
cd nginx/gui
yarn install
yarn start
```

Run backend server

```
# Set environment vars
source ./env/dev.sh 

# Initialize python virtual env
cd django
python3 -m venv env
source env/bin/activate
pip3 install -r requirements.txt

# Initialize database
./init_db.sh
python3 manage.py runserver
```

## 🎈 Script Usage <a name="usage"></a>

```
# Authenticate with GKE cluster
./bin/auth.sh 

# Ensure env variables are set
./bin/check_env.sh

# Helm script
./bin/helm.sh install|upgrade|delete dev|stage|prod

# Init/get/push/delete secrets to GCP
./bin/secrets.sh init|get|push|delete

# Create/delete CI/CD triggers
./bin/triggers.sh create|delete|reset
```

## ⛏️ Built Using <a name = "built_using"></a>

- [Django Rest Framework](https://www.django-rest-framework.org/) - Python Rest API framework
- [ReactJS](https://reactjs.org/) - Typescript Frontend library
- [Cloud SQL](https://https://cloud.google.com/sql) - Google Managed Database
- [Docker](https://www.docker.com/) - Build Container Images
- [Kubernetes](https://kubernetes.io/) - Cloud Environment
- [Helm](https://helm..sh/) - Kubernetes Deployment
- [Terraform](https://terraform.io/) - Cloud IAC -> repo found [here](https://github.com/walkerobrien/lakegames-infra.git/)
- [Google Cloud Platform](https://www.cloud.google.com/) - Public Cloud

## 🚀 CI/CD <a name = "cicd"></a>
The following branches are connected to [Google Cloud Build](https://console.cloud.google.com/cloud-build/builds?project=lakegames-gke-prod) CI/CD pipelines 
- `dev`
- `stage`
- `master`

## ✍️ Authors <a name = "authors"></a>

- [@walkerobrien](https://github.com/walkerobrien) 
  - Project Manager
  - Lead Developer
  - Cloud Architect
- [@meredithkeesling](https://github.com/meredithkeesling) 
  - Frontend Web Designer



