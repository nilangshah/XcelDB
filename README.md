XcelDB
======

Keystore written in GO with Raft consensus. XcelDB is highly available, scalable, key value store. XcelDB follow ACID property. It automatically replicate data to cluster and store it on disk. It uses Raft consensus protocoal to provide high availablity and strong consistency.  



Features
====

* Distributed
* Strongly consistent
* Highly Available
* scalable
* 

Requirements
============
* Need alteast go1.1 or newer
* Need Raft
(see usage for more info)

Usage
===== 
* Install Raft from github.com/nilangshah/Raft (see usage instruction of Raft)
* Run tests to test XcelDB 
 * Make

* To change number of server in cluster make changes in config.html and c_config.html
 * config.xml is for http front-end xceldb and c_config.xml is for back-end cluster server communication.
