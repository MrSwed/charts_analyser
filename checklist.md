checklist:
- [x] migration and import data
- [x] postgis, docker
- [x] app, handlers
  - [x] get vessels for zone GET /vessels?
  - [x] get zones for vessels GET /zones?
  - [ ] set monitoring mode POST /monitor
  - [ ] store track POST /track
  - [ ] monitoring GET /monitor  

    [//]: # (  - [ ] track log GET /track )
- [ ] vessel client (simulator)
   - [ ] repository get tracks by vessel_id 
- [x] redis, docker
- [ ] operator set vessel to monitoring (only status: control)
  - [ ] option: "allow set automatically from vessel"
  - [ ] save start/finish monitored time (postgres table `state_log`:  
        `id, timestamp, vessel_id, state(control/sleep/awaiting control), comment`
  - [ ] option "monitoring time out" time, after that vessel will be removed
        from monitoring if no new data from vessel (redis), 0 - only by operator
- [ ] receive data from vessel 
     - [ ] store to Redis if vessel at monitored status 
      a vessel key in the redis, means it at monitored status  (redis db 0)
     - save to track_log, postgres
- [ ] online monitoring (from redis), give data (for operator), handlers:
  - [ ] list of controlled vessels (list of keys of deb_0)
  - [ ] monitored vessel (timestamp, status, point(log,ltd), current map id, time spent in map)
  - [ ] 
- [ ] auth as middleware for roles: vessel, operator
- 
- [ ] app to docker, with build layers and individual containers for app and simulator
