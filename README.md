# Server Sent Events
Implementation of SSE in go

## Project Structure
The main part of this project is an Event Service which is defined in `svc/svc.go` file. It is constructed out of a service that is able to create Topics, which are responsible of tracking and informing Subscriptions. When some caller is publishing new data to the service, an event is created and then passed to a related topic. The topic then iterates over it's subscribers and pushes the event to a subscribers event channel.
When a caller wants to subscribe to a topic it calls the `Subscribe` method provided by the event service. This method returns a Subscription object that is used to receive events and unsubscribe from the topic (by closing the subscription).

## Http API
This project is using Fiber http framework to define and serve event service related endpoints.
At the moment there are only two endpoints, one for subscription antoher one for publishing an event:
* GET /:topic - subscribers to the topic and waits for events any new events. 
* POST /:topic - publishes arbitrary data. No error is returned if there are no subscribers to a given topic.

Default subscribtion timeout for a topic is 30 seconds. 
