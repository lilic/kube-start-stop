# kube-start-stop

Schedule Scaling of Kubernetes Resources.

Automatically start and stop Kubernetes resources in the namespace. Schedule your resources to automatically scale down during a desired time period, e.g. scale down your dev workloads during the weekend.

# Usage

Deploy controller:

```
  kubectl apply -f manifests/ 
```

Example manifest file:
```
apiVersion: scheduler.io/v1alpha1
kind: Schedule
metadata:
  name: schedule
  labels:
    schedule: weekly-dev
spec:
  schedules:
  - replicas: 0
    selector: my-deployment
    start:
      day: Monday 
      time:
        hour: 16 
        minute: 30
    stop:
      day: friday
      time:
        hour: 8
        minute: 10
```

After filling out the manifest file with the desired schedule, you just have to apply that your cluster that is running the kube-start-stop controller:
```
  kubectl apply -f example.yml
```
