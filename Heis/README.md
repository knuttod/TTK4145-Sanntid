# Elevator project TTK4145

## Running the program

An elevatorServer or simElevatorServer must be used. These were given in the project and the sim can be found at [github](https://github.com/TTK4145)

The code can be runned by typing

```console
go run main.go -port=<PORT>
```
Alternatively some more flags can be used

```console
-id=<id>
-processPairsFlag=<true/false>
```

The port flag must be used and the port specified is the one that the program should communicate with the sim-/elevatorserver on. If no id is given a random id will be assigned. The last flag is to enable a processPairs backup system, this is by default turned off as this functionality only work on Linux operating systems. 

## Description
Our system can in large be divided into three main modules. 


### FSM
A state machine for a single elevator. It interacts with the elevetor io, which in itself is a standalone module, by getting information about button presses and current floor and setting lights and motor direction. Button presses are sent to the orders module for assignment of orders and eventual orders to be taken by the fsm are given from the orders module. 

### Orders

The order module is responsible for handling and synchronising all orders, either comming from a local button press or from updates on the network. All elevators on the network keeps track of the other elevators orders in a map called AssignedOrders, where the keys are elevator id's and the values are a 2d slice of assigned orders for the corresponding elevator implemented as a cyclic counter. The module assigns orders to the most cost efficient elevator and orders are reassigned when an elevetor either disconnects or are unable to take its own orders (motorstop/obstruction). When an order is to be done by a local elevator this module sends this to the FSM

### Network

The network module is responsible for sending information to the other elevators brodcasting using UDP. Each elevator sends information about itself and all it knows about the orders of itself and the other elevators. This together with its id is sent many times a second which also makes the module able to keep track of elevators disconnecting and connecting. An elevator is assumed disconnected when others have not received a message from it in a while. This should also handle packet loss. The elevator states is JSON enoded.

## Config file

In the config file parameters as number of elevators, number of floors, how long the door should be open, etc. can be changed. 
