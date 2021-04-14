# 1. Create a cluster
First choose 3 nodes, one is master and other two are workers. All the nodes we use ubuntu OS. The master node and the worker nodes should install the following components.

| IP            | Hostname | Componets                                                                                          | info                            |
| ------------- | -------- | -------------------------------------------------------------------------------------------------- | ------------------------------- |
| 10.148.188.77 | master   | kube-apiserver, kube-controller-manager, kube-scheduler, etcd, kubelet, docker, flannel, dashboard | RAM 16GB,VCPUs:8 VCPU;Disk:80GB |
| 10.148.188.74 | node1    | kubelet, docker, flannel、traefik                                                                  |
| 10.148.188.76 | node2    | kubelet, docker, flannel                                                                           |

## 1.1 Install critical components for all nodes
### 1.1.1 install container-runtime(docker)
we install [docker](https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository) as the basic container runtime.
```
sudo apt-get update
sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
sudo apt-get update
sudo apt-get install docker-ce=5:19.03.14~3-0~ubuntu-focal docker-ce-cli=5:19.03.14~3-0~ubuntu-focal containerd.io # install stable docker
```

### 1.1.2 install kubeadm, kubelet and kubectl
we install 1.20 release kube-apiserver into our master node.

firstly, we should install binary: [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/). According to the mannual:
- Letting iptables see bridged traffic
```
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo sysctl --system
```
- install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) and related operations
```
curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
kubectl version --client
```
The result is 
```
root@minghyuan-master-4141994:~# kubectl version --client
Client Version: version.Info{Major:"1", Minor:"20", GitVersion:"v1.20.1", GitCommit:"c4d752765b3bbac2237bf87cf0b1c2e307844666", GitTreeState:"clean", BuildDate:"2020-12-18T12:09:25Z", GoVersion:"go1.15.5", Compiler:"gc", Platform:"linux/amd64"}
```
It seems very well.

- install kubeadm
```
sudo apt-get update && sudo apt-get install -y apt-transport-https curl
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```
The kubelet is now restarting every few seconds, as it waits in a crashloop for kubeadm to tell it what to do. Then restart the kubelet:
```
sudo systemctl daemon-reload
sudo systemctl restart kubelet
```

## 1.2 init master node
1. init api-server
```
rm -rf /etc/kubernetes/
echo 1 > /proc/sys/net/ipv4/ip_forward
systemctl stop kubelet
kubeadm init --apiserver-advertise-address $(hostname -i) --pod-network-cidr=10.0.0.0/16
```
The following messages are very useful:
```
To start using your cluster, you need to run the following as a regular user:
```
  mkdir -p $HOME/.kube
  rm $HOME/.kube/config
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config
```
Alternatively, if you are the root user, you can run:

  export KUBECONFIG=/etc/kubernetes/admin.conf

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

Then you can join any number of worker nodes by running the following on each as root:

kubeadm join 10.148.188.77:6443 --token 2x0ql3.4h6z2rb8hc1d9mfy \
    --discovery-token-ca-cert-hash sha256:211a22560d71c74937a01e55e0088c38797562fa01783fd2270118ec9c766065
```


2. edit the kubeconfig
```
  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config
```
3. we use the ["Flannel"](https://github.com/coreos/flannel/blob/master/Documentation/kubernetes.md) as our network policy

```
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/k8s-manifests/kube-flannel-rbac.yml
```

4. worker nodes add master node.
My master node ip is 10.148.188.77, according the kubeadm's information, we can exec the following commands to add the master:
```
kubeadm join 10.148.188.77:6443 --token 2x0ql3.4h6z2rb8hc1d9mfy \
    --discovery-token-ca-cert-hash sha256:211a22560d71c74937a01e55e0088c38797562fa01783fd2270118ec9c766065
```
the workers' messages are
```
This node has joined the cluster:
* Certificate signing request was sent to apiserver and a response was received.
* The Kubelet was informed of the new secure connection details.

Run 'kubectl get nodes' on the control-plane to see this node join the cluster.
```

in work nodes:
```
root@minghyuan-node2-4141025:~# kubectl get nodes
The connection to the server localhost:8080 was refused - did you specify the right host or port?
```

if you want to add more nodes in this cluster, you can repeat exec ##1.2 step4.

## 1.2 trouble shooting

in master node:
```
root@minghyuan-master-4141994:~# kubectl get nodes
NAME                       STATUS     ROLES                  AGE    VERSION
minghyuan-master-4141994   NotReady   control-plane,master   21m    v1.20.1
minghyuan-node1-4140739    NotReady   <none>                 2m     v1.20.1
minghyuan-node2-4141025    NotReady   <none>                 117s   v1.20.1
```
It means that in worker nodes, the port 8080 are closed, and in master nodes, we exec kubectl to call 8080 port to the master nodes' local api-server, so we can use 'kubectl ' commands from master nodes to get information.

The reason of nodes' status are "NotReady" are because the nodes have taints. And the important component Coredns cannot be started in master nodes. 
```
  taints:
  - effect: NoSchedule
    key: node.kubernetes.io/not-ready
```

after that we get all the pods:
```
root@minghyuan-master-4141994:~# k get pod  --all-namespaces
NAMESPACE     NAME                                               READY   STATUS              RESTARTS   AGE
kube-system   coredns-74ff55c5b-8srb2                            0/1     ContainerCreating   0          18h
kube-system   coredns-74ff55c5b-vwbgd                            0/1     ContainerCreating   0          18h
kube-system   etcd-minghyuan-master-4141994                      1/1     Running             0          18h
kube-system   kube-apiserver-minghyuan-master-4141994            1/1     Running             0          18h
kube-system   kube-controller-manager-minghyuan-master-4141994   1/1     Running             0          18h
kube-system   kube-proxy-hcvbt                                   1/1     Running             0          18h
kube-system   kube-proxy-qv5s2                                   1/1     Running             0          17h
kube-system   kube-proxy-tzlzm                                   1/1     Running             0          17h
kube-system   kube-scheduler-minghyuan-master-4141994            1/1     Running             0          18h
```
you can see [here](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/troubleshooting-kubeadm/),

## 1.3 intall CNI in master node

we use [flannel](https://github.com/coreos/flannel#flannel) as our network CNI

```
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
```
modify /etc/docker/daemon.json change cgroudfs into systemd according to the kubernetes's recommended.

```
vi /etc/docker/daemon.json

{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
```


```
mkdir -p /etc/cni/net.d
cat >/etc/cni/net.d/10-mynet.conf <<EOF
{
    "cniVersion": "0.2.0",
    "name": "mynet",
    "type": "bridge",
    "bridge": "cni0",
    "isGateway": true,
    "ipMasq": true,
    "ipam": {
        "type": "host-local",
        "subnet": "10.22.0.0/16",
        "routes": [
            { "dst": "0.0.0.0/0" }
        ]
    }
}
EOF

cat >/etc/cni/net.d/99-loopback.conf <<EOF
{
    "cniVersion": "0.2.0",
    "name": "lo",
    "type": "loopback"
}
EOF
sudo systemctl restart kubelet
```
## 1.4 fix the worker nodes "notReady" problem
```
root@minghyuan-master-4141994:~# k get node
NAME                       STATUS     ROLES                  AGE   VERSION
minghyuan-master-4141994   Ready      control-plane,master   20h   v1.20.1
minghyuan-node1-4140739    NotReady   <none>                 20h   v1.20.1
minghyuan-node2-4141025    NotReady   <none>                 20h   v1.20.1
root@minghyuan-master-4141994:~# k get pod --all-namespaces
NAMESPACE     NAME                                               READY   STATUS             RESTARTS   AGE
kube-system   coredns-6ccb5d565f-4k48p                           1/1     Running            0          4h56m
kube-system   coredns-6ccb5d565f-mgzc4                           1/1     Running            0          3h9m
kube-system   etcd-minghyuan-master-4141994                      1/1     Running            0          23h
kube-system   kube-apiserver-minghyuan-master-4141994            1/1     Running            0          23h
kube-system   kube-controller-manager-minghyuan-master-4141994   1/1     Running            0          23h
kube-system   kube-flannel-ds-4xsd9                              0/1     CrashLoopBackOff   2          34s
kube-system   kube-flannel-ds-g9rjn                              0/1     Pending            0          43m
kube-system   kube-flannel-ds-kzljz                              0/1     Pending            0          42m
kube-system   kube-proxy-hcvbt                                   1/1     Running            0          23h
kube-system   kube-proxy-qv5s2                                   1/1     Running            0          23h
kube-system   kube-proxy-tzlzm                                   1/1     Running            0          23h
kube-system   kube-scheduler-minghyuan-master-4141994            1/1     Running            0          23h
root@minghyuan-master-4141994:~# k logs kube-flannel-ds-4xsd9 -n kube-system
ERROR: logging before flag.Parse: I1230 09:34:18.554556       1 main.go:519] Determining IP address of default interface
ERROR: logging before flag.Parse: I1230 09:34:18.554980       1 main.go:532] Using interface with name eth0 and address 10.148.188.77
ERROR: logging before flag.Parse: I1230 09:34:18.555003       1 main.go:549] Defaulting external address to interface address (10.148.188.77)
W1230 09:34:18.555026       1 client_config.go:608] Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.
ERROR: logging before flag.Parse: I1230 09:34:18.758062       1 kube.go:116] Waiting 10m0s for node controller to sync
ERROR: logging before flag.Parse: I1230 09:34:18.758480       1 kube.go:299] Starting kube subnet manager
ERROR: logging before flag.Parse: I1230 09:34:19.846098       1 kube.go:123] Node controller sync successful
ERROR: logging before flag.Parse: I1230 09:34:19.846170       1 main.go:253] Created subnet manager: Kubernetes Subnet Manager - minghyuan-master-4141994
ERROR: logging before flag.Parse: I1230 09:34:19.846177       1 main.go:256] Installing signal handlers
ERROR: logging before flag.Parse: I1230 09:34:19.846377       1 main.go:391] Found network config - Backend type: VXLAN
ERROR: logging before flag.Parse: I1230 09:34:19.846487       1 vxlan.go:123] VXLAN config: VNI=1 Port=0 GBP=false Learning=false DirectRouting=true
ERROR: logging before flag.Parse: E1230 09:34:19.846857       1 main.go:292] Error registering network: failed to acquire lease: node "minghyuan-master-4141994" pod cidr not assigned
ERROR: logging before flag.Parse: I1230 09:34:19.846946       1 main.go:371] Stopping shutdownHandler...
```

reinstall flennel

```
# delete the old flannel daemonSet
kubectl delete ds kube-flannel-ds -n kube-system

# create a new flannel daemonSet
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
```

Check the events:
```
kubectl describe pod -n kube-system kube-flannel-ds-m5w4p
  ----     ------     ----                  ----               -------
  Normal   Scheduled  21m                   default-scheduler  Successfully assigned kube-system/kube-flannel-ds-m5w4p to minghyuan-master-4141994
  Normal   Pulled     21m                   kubelet            Container image "quay.io/coreos/flannel:v0.13.1-rc1" already present on machine
  Normal   Created    21m                   kubelet            Created container install-cni
  Normal   Started    21m                   kubelet            Started container install-cni
  Normal   Pulled     20m (x5 over 21m)     kubelet            Container image "quay.io/coreos/flannel:v0.13.1-rc1" already present on machine
  Normal   Created    20m (x5 over 21m)     kubelet            Created container kube-flannel
  Normal   Started    20m (x5 over 21m)     kubelet            Started container kube-flannel
  Warning  BackOff    6m37s (x70 over 21m)  kubelet            Back-off restarting failed container
  Normal   Pulled     4m3s (x4 over 5m41s)  kubelet            Container image "quay.io/coreos/flannel:v0.13.1-rc1" already present on machine
  Normal   Created    4m3s (x4 over 5m41s)  kubelet            Created container kube-flannel
  Normal   Started    4m2s (x4 over 5m41s)  kubelet            Started container kube-flannel
  Warning  BackOff    31s (x24 over 5m39s)  kubelet            Back-off restarting failed container
```
check the logs:

```
root@minghyuan-master-4141994:~# k logs -n kube-system kube-flannel-ds-m5w4p
ERROR: logging before flag.Parse: I0104 06:40:42.347579       1 main.go:519] Determining IP address of default interface
ERROR: logging before flag.Parse: I0104 06:40:42.348127       1 main.go:532] Using interface with name eth0 and address 10.148.188.77
ERROR: logging before flag.Parse: I0104 06:40:42.348153       1 main.go:549] Defaulting external address to interface address (10.148.188.77)
W0104 06:40:42.348173       1 client_config.go:608] Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.
ERROR: logging before flag.Parse: I0104 06:40:42.545957       1 kube.go:116] Waiting 10m0s for node controller to sync
ERROR: logging before flag.Parse: I0104 06:40:42.546016       1 kube.go:299] Starting kube subnet manager
ERROR: logging before flag.Parse: I0104 06:40:43.546182       1 kube.go:123] Node controller sync successful
ERROR: logging before flag.Parse: I0104 06:40:43.546217       1 main.go:253] Created subnet manager: Kubernetes Subnet Manager - minghyuan-master-4141994
ERROR: logging before flag.Parse: I0104 06:40:43.546222       1 main.go:256] Installing signal handlers
ERROR: logging before flag.Parse: I0104 06:40:43.546345       1 main.go:391] Found network config - Backend type: vxlan
ERROR: logging before flag.Parse: I0104 06:40:43.546432       1 vxlan.go:123] VXLAN config: VNI=1 Port=0 GBP=false Learning=false DirectRouting=false
ERROR: logging before flag.Parse: E0104 06:40:43.546853       1 main.go:292] Error registering network: failed to acquire lease: node "minghyuan-master-4141994" pod cidr not assigned
ERROR: logging before flag.Parse: I0104 06:40:43.546935       1 main.go:371] Stopping shutdownHandler...
```

so i want to reset the worker nodes, in master nodes:
```
kubeadm reset # and delete the related dirs
```

it says:
```
[ERROR Port-10250]: Port 10250 is in use
```
check what port are token:
```
netstat -tunlp|grep 10250
```

kubeadm join 10.148.188.77:6443 --token bkb2nb.66226fkzzplwb2nm \
    --discovery-token-ca-cert-hash sha256:8a99cabf0265536f03703cb89b3cadfd0478ff40cd0f4b1acce001fd95782ce6