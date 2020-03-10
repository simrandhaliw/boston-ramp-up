# Operator Project

## Assignment
Create a go based operator using the Operator SDK https://github.com/operator-framework/operator-sdk
This operator should be based around the TimeServer CRD which you will create.
The controller for TimeServer should create a HTTP server that returns the current time when queried.
One server should be created for each replica specified in the CRD. Each TimeServer should exist on a different
worker node. If there are more replicas specified than worker nodes, only create as many TimeServers as nodes.
When the amount of replicas in the CRD is changed, the controller should change the number of TimeServers if possible.

## Bonus
Watch the node objects in the cluster as well. When the number of nodes change, adjust the amount deployed TimeServers
if necessary

## Super Stretch Bonus
Write end to end tests to test the functionality of your operator. Your tests should alter the TimeServer replicas
and ensure that the actual TimeServers reflect that change. You should do the same with the node count. You should
also ensure that each node has exactly 1 or 0 TimeServers deployed on it.
