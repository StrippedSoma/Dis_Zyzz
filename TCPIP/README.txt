a) "What are packages in your implementation? What data structure do you use to transmit data and meta-data?"
    Our packages are integers sent by channels between our server- and client threads.

b) "Does your implementation use threads or processes? Why is it not realistic to use threads?"
    Our implementation utilizes threads, but that is not feasible for real use as threads cannot be split into smaller units of execution.
    What this means is, that we cannot multithread the communication.

c) "In case the network changes the order in which messages are delivered, how would you handle message re-ordering?"
    We have a counter 'attempts' that will reset the loop in case we receive unexpected responses.

d) "In case messages can be delayed or lost, how does your implementation handle message loss?"
    When our threads send a message, they start a timer. If they do not receive a response in a given time period,
    the message is assumed lost or delayed and communication is restarted.

e) ""Why is the 3-way handshake important?"
    The three-way handshake in TCP serves two key purposes. 
    The primary purpose is that to make sure that both parties are prepared to exchange data. 
    Furthermore it ensures the agreement on initial sequence numbers.