# unique-devices

A scaffolding for building Go services at the Wikimedia Foundation in order to establish similar patterns amongst all Go services.

### Docker Quickstart

Generated a Dockerfile for service variant with `blubber .pipeline/blubber.yaml <variant> > Dockerfile`,
and build using regular Docker tools.


For example, build and run a `development` variant of a service with:
```
blubber .pipeline/blubber.yaml development > Dockerfile
docker build -t unique-devices .
docker run -p 8000:8000  unique-devices
```

Connect to `http://localhost:8000/healthz`.
