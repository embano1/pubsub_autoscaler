# Difference between autoscaler and autoscaler_ext
autoscaler_ext is meant for quick testing of new functions and features from outside the Kubernetes cluster, i.e. 
don´t have to build a new Docker image after every small change.  

In case you want to play with autoscaler_ext at least modify  
kubeconfig = flag.String("kubeconfig", "/Users/mgasch/.kube/config", "absolute path to the kubeconfig file")  
to match your kube config file (autoscaler_ext will use the current kube context, so make sure to target the right cluster before using it).  

The build process (Makefile) uses autoscaler since this will be a pod to be run inside the cluster.
