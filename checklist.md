checklist:
- [x] migration and import data
- [x] postgis, docker
- [x] app, handlers
  - [x] get vessels for zone GET /vessels?zone_name=name
  - [x] get zones for vessels GET /zones?vessel_id=XXvessel_id=XX
  - [x] set monitoring mode POST /monitor/?vessel_id=XX
  - [x] monitoring GET /monitor - list of monitored vessels  
  - [x] get monitored vessel info GET /monitor/state?vessel_id=XX
  - [x] store track POST /track/vessel_id=XX && update monitored

      [//]: # (  - [ ] track log GET /track )
- [x] vessel client (simulator)
   - [x] repository get tracks by vessel_id for new online data
- [x] redis, docker
- [x] operator set vessel to monitoring
    - [x] redis controlled list collection

       [//]: # (  - [ ] option: "allow set automatically from vessel")
  - [x] save start/finish monitored time (postpone postgres table `control_log`:  
        `id, timestamp, vessel_id, state(control/sleep/awaiting control), comment`

      [//]: # (  - [ ] option "monitoring time out" time, after that vessel will be removed)
        from monitoring if no new data from vessel (redis), 0 - only by operator
- [x] receive data from vessel 
     - [x] store to Redis if vessel at monitored status 
     - [x] save postpone to tracks, postgres
- [x] online monitoring (from redis), give data (for operator), handlers:
  - [x] list of controlled vessels (list of keys of deb_0)
  - [x] monitored vessel (`timestamp, status, point(log,ltd), current map id, time spent in map`)
- [x] auth as middleware for roles: vessel, operator

- [ ] app to docker, with build layers and individual containers for app and simulator
