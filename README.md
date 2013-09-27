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

Configuration for each source is stored in a config.json file. If the configuration 
doesn't exist, the source won't be loaded. Below is a sample configuration file. 
I group each of my sources into different Choir channels, but that's not necessary.
```json
{
  "sources": [
    {
      "type": "jira",
      "key": "choirkey1",
      "http": {
        "domain": "xxx.jira.com",
        "username": "basic",
        "password": "auth"
      }
    },
    {
      "type": "github",
      "key": "choirkey2",
      "http": {
        "username": "dacort",
        "orgname": "myorg",
        "access_token": "github_oauth_token"
      }
    },
    {
      "type": "campfire",
      "key": "choirkey3",
      "rooms": [1,2,3],
      "token": "campfire_token",
      "orgname": "myorg"
    },
    {
      "type": "yammer",
      "key": "choirkey4",
      "http": {
        "access_token": "yammer_access_token"
      }
    },
    {
      "type": "desk",
      "key": "choirkey4",
      "http": {
        "orgname": "myorg",
        "username": "damon@myorg.com",
        "password": "auth"
      }
    }
  ]
}
```