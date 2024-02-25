checklist:
- [x] migration and import data
- [x] postgis, docker
- [x] app, handlers
  - [x] get vessels for zone GET /vessels?zone_name=name
  - [x] get zones for vessels GET /zones?vessel_id=XXvessel_id=XX
  - [x] set monitoring mode POST /monitor/:id

      [//]: # (  - [ ] set monitoring mode bath POST /monitor/)
  - [x] monitoring GET /monitor - list of monitored vessels  
  - [ ] get monitored vessel info GET /monitor/:id
  - [ ] store track POST /track/vessel_id=XX && update monitored

      [//]: # (  - [ ] track log GET /track )
- [ ] vessel client (simulator)
   - [ ] repository get tracks by vessel_id for new online data
- [x] redis, docker
- [ ] operator set vessel to monitoring (only status: control)
    - [ ] redis controlled list collection

       [//]: # (  - [ ] option: "allow set automatically from vessel")
  - [ ] save start/finish monitored time (postpone postgres table `state_log`:  
        `id, timestamp, vessel_id, state(control/sleep/awaiting control), comment`
  - [ ] option "monitoring time out" time, after that vessel will be removed
        from monitoring if no new data from vessel (redis), 0 - only by operator
- [ ] receive data from vessel 
     - [ ] store to Redis if vessel at monitored status 
     - [ ] save postpone to track_log, postgres
- [ ] online monitoring (from redis), give data (for operator), handlers:
  - [ ] list of controlled vessels (list of keys of deb_0)
  - [ ] monitored vessel (`timestamp, status, point(log,ltd), current map id, time spent in map`)
  - [ ] 
- [ ] auth as middleware for roles: vessel, operator
- 
- [ ] app to docker, with build layers and individual containers for app and simulator
