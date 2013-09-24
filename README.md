Choirmaster is a tool to poll various data sources and make them sing with choir.io.

Overview
########

Choirmaster is a fun and easily extensible framework for plugging various data sources
into choir.io. Each data source controls it's own destiny, but the intent is that they
poll their remote API/site/feed every X minutes and construct a Choir Note for every 
event. 

Usage
#####


Each module (member of the ensemble) must register themselves with an init() method and
implement at least 2 methods:
 - Configure
 - Run

Configure takes a config interface and knows how to configure itself using that blog.
Run Will do whatever polling of the remote data source is necessary and send any 
Choir Notes back to the main process via the passed channel.