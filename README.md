[![Build Status](https://cloud.drone.io/api/badges/jones2026/drone-flowdock/status.svg)](https://cloud.drone.io/jones2026/drone-flowdock)
[![](https://images.microbadger.com/badges/image/jones2026/drone-flowdock.svg)](https://microbadger.com/images/jones2026/drone-flowdock "Get your own image badge on microbadger.com")

# drone-flowdock
Drone plugin to push messages to Flowdock

#### Image name:
`jones2026/drone-flowdock`

#### Settings

| setting | required | description |
------------- | ------------- | ----------
flow_token | yes | Flowdock token for flow that message will be posted to. [Steps to create token can be found here.](docs/flowdock-setup.md)
message | yes | Message that will be posted to Flowdock.
files | no | Specify file or pattern of files to be uploaded to the same thread as the message posted.
max_files | no | Defaults to 5. This is to ensure the flow is not flooded if the file pattern matches too many files.

#### Example usage

```
- name: flowdock
  image: jones2026/drone-flowdock
  settings:
      message: ":red_circle: failure on Drone :point_right: ${DRONE_BUILD_LINK}"
      flow_token:
          from_secret: FLOWDOCK_TOKEN
  when:
      status:
          - failure
```
