# Storage proxy workflow

- logic abstract

> Mainly adopt the observer mode to design the storage proxy

| Observer role | Interpretation | Object mapped in business |
|:-:|:-|:-|
| Event | The object of the observed person, corresponding to the event data model in the business | such as returning the information of order expiration |
| Observer | Observer, corresponding to the caller in the business, which can obtain the event data changes observed by it in real time | There are two types of business that can be mapped to this role <br> 1 Automatic renewal module <br> 2 Third-party callback |
| Notifier | controller, initiator of event notification | monitoring module in storage proxy |
