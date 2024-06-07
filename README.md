# Chandy-Lamport distributed algorithm
The Chandy-Lamport distributed algorithm is a famous algorithm to record a global consistent snapshot of a distributed system/application:<br>
a snapshot records the local state of each process along with the state of each communication channel used by the processes to communicate.

```@article{ChandyLamportDistributedAlgorithm,
Reference article:
• title={Distributed Snapshots: Determining Global States of a Distributed System},
• authors={Leslie Lamport, K. Mani Chandy},
• journal={ACM Transactions on Computer Systems},
• volume={3},
• number={1},
• pages={63-75},
• year={1985}
```


## Properties
This project is about design, implementation, and evaluation of the Chandy-Lamport algorithm for snapshotting the global state of a distributed system.
The solution is tested on a pipelined deployed application that works like this: there are processes that all start with the same balance (in dollars); every second, each process transfers funds to another process at random; a process, also chosen at random, takes a snapshot of the system every two seconds; this whole process is carried out for a total of 10 seconds; funds transferred and snapshots taken are displayed to the user. Ultimately, an `output` directory will be created with network process logs and GoVector logs, which can be useful for monitoring, verifying, diagnosing, and replaying distributed system states.<br>
Then, this program calculates a snapshot of financial transactions based on the Chandy-Lamport algorithm.

For a more detailed specification of the system, please see the report.


## Dependencies
This project use [GoVector](https://github.com/DistributedClocks/GoVector) for drawing the trace of the network messages sent 
among the nodes to perform the global snapshot.


## Setup


## Usage
