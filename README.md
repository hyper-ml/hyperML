
**hyperflow** is a Machine learning platform for growing teams. The platform adds reusuability and scalability to your machine learning stack. It does by componentizing model development and executing training tasks inside containers.

You can use the current version of hflow just keeping track of model-experiment cycle or also utilize containerized (docker) training and automated version control when deployed with kubernetes.

The project is in alpha stage but if you are keen to try out do write us at beekeepr@hyperflow.cc

![Alt Text](http://www.animatedgif.net/underconstruction/cns01_e0.gif)

### Key Features:

* Reduced experimentation time: Componentization and parallel training on cloud to optimise time to experiment. Don't spend weeks to experiment when you can finish in days.
* Ease-of-use: Launch tasks (e.g training, data processing) with a single command
* Version control: Code, data and model versioning for reproducibility
* Track experiments: Keep track of experiment results with versions. Whether you are tuning hyperparameters, or trying out variants of a model, with hyperflow you don't need to queue up training jobs on your personal system.
* Deploy on premise or on private cloud
* Team Collaboration: Clone repos to share work with team members, share datasets, resources and Infrastructure
* Documentation: Write documentation around repo, model or dataset usage
* Framework Support: Tensorflow, Keras, pytorch out of box. You can add your own as a docker image
* GPU Support

### Documentation
https://hyperflow.in/docs/introduction

### Example
A typical workflow with hflow will look like: 

##### 1. Initiate repo
  $ hflow init my_repo <br />
  > Repo my_repo initiated

##### 2. upload dataset (optional) 
  $ cd /home/data  <br />
  > hflow init my_dataset  <br />
  > hflow push <br />

##### 3. submit training task   
  $ cd my_repo <br />
  $ hflow run "python my_training_program.py"  --data my_dataset <br />
  > Flow Id: erwed43i5jin5423423d <br />

##### 4. monitor status 
  $ hflow status  <br />
  > Flow Id: erwed43i5jin5423423d <br />
  > Status: RUNNING <br />

##### 5. view execution log 
  $ hflow log <br />

##### 6. download saved model (optional)
  $ hflow pull saved_models <br />
  > Model files downloaded to ./saved_models

##### 7. download results or output files (optional)
  $ hflow pull results  <br />
  > Results downloaded in ./results <br />


