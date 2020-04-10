# hyperML
Radically simple, efficient platform to abstract infrastructure complexities from data scientists and ML engineers. 

hyperML works on top of kubernetes to provide reusable environments, launch or schedule notebooks and python jobs. If you are sharing resources then you can queue up requests as well.

Demo Links: 
+ [Launch Notebook](https://wizardly-davinci-f7fc85.netlify.com/demo.mp4)  
+ [CLI](https://wizardly-davinci-f7fc85.netlify.com/hflow.gif)  
+ [Schedule Notebooks from Jupyter Labs](https://wizardly-davinci-f7fc85.netlify.com/schedule-notebook.gif)  

## Problems solved
* Ease of use (follows general dialect of data science)
* Abstracts infrastructure from data scientists and ML engineers leting them focus more on science
* Extend your local environment with on-demand cloud resources without having to leave your favourite IDE (jupyter labs at the moment but VS code extension is planned).
* Scale ML experiments by effortlessly launching new notebooks or scheduling them to run in background
* Share infrastructure resources especially when there is shortage of it 
* Hassle free environments through container images


## Requirements
* Kubernetes (minikube, on-premise, AWS EKS or GKE or any public cloud) 

## Getting Started
Install standalone binary 

```
curl -LO curl http://storage.googleapis.com/hyperml/releases/0.9.0/hyperml /usr/local/bin/hyperml
```

You can also install hyperML as lambada function to optimize server costs.

A host of quick start configuration guides are available on hyperML [website](https://www.hyperml.com/docs/prerequisites)
* [Local install](https://www.hyperml.com/docs/standalone)
* [Minikube Install](https://www.hyperml.com/docs/minikube)
* [Kubernetes Install](https://www.hyperml.com/docs/kubernetes)
* [AWS EKS](https://www.hyperml.com/docs/aws-eks)
* [GCP GKE](https://www.hyperml.com/docs/gcp-gke)

## User flow:
* Run or Schedule **Notebooks to run in the background** right from the comform of jupyter labs  
* **Launch notebooks** on click of a button  
* Install **locally** or on remote server (VM) or as a container inside kubernetes
* Simple CLI to run Python code bundles inside containers on kubernetes 


## Documentation
A comprehensive documentation is available [here](https://www.hyperml.com/docs/introduction) link

## Problems?
If you have a problem, please raise an issue. 

## Author
Amol Umbarkar (amol@hyperml.com / [twitter](https://twitter.com/_4mol))


