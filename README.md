go_blog
gin + vue3 开发的个人博客项目

本地测试
开发工具及版本
golang: 1.23.5

node: v22.13.0

docker: 27.4.0

编译器：vscode


单节点K8S部署

1.1 查看cpu信息 k8s安装至少需要2核4G的环境，否则会安装失败

```bash
lscpu
```

1.2 安装k8s时，临时关闭swap ，如果不关闭在执行kubeadm部分命令会报错

```bash
swapoff -a  *# 临时* 
sed -ri 's/.*swap.*/#&/' /etc/fstab    *# 永久*
```

1.3 安装k8s时，可以临时关闭selinux，减少额外配置
```bash
sed -i 's/enforcing/disabled/' /etc/selinux/config  *# 永久* 
setenforce 0  *# 临时*
```

1.4 关闭防火墙
```bash
systemctl stop firewalld
systemctl disable firewalld
```
1.5 设置网桥参数
```bash
cat << EOF > /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
```
```bash
sysctl --system  *# 生效*
```
1.6 修改hosts文件 方便查看域名映射
```bash
vim /etc/hosts
```
追加以下内容, 将IP更换为真实IP
```bash
X.X.X.X master
```
让修改生效
```bash
systemctl restart network
```
1.7 修改hostname
```bash
hostnamectl set-hostname master
```
2.1 安装软件
```bash
yum install -y yum-utils device-mapper-persistent-data lvm2
```
2.2 安装docker:
```bash
wget https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo -O /etc/yum.repos.d/docker-ce.repo
yum -y install docker-ce
systemctl enable docker && systemctl start docker
```
2.3 配置镜像下载加速器：
```bash
vim /etc/docker/daemon.json
{
    "registry-mirrors": [
        "https://docker.m.daocloud.io",
        "https://docker.nju.edu.cn",
        "https://dockerproxy.com",
        "https://docker.udayun.com",
        "https://noohub.ru",
        "https://huecker.io",
        "https://dockerhub.timeweb.cloud"
    ]
}
systemctl restart docker
docker info     #查看docker信息，进行确认
```
2.4 添加阿里云软件源：
```bash
cat > /etc/yum.repos.d/kubernetes.repo << EOF
[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=0
repo_gpgcheck=0
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF

cat > /etc/yum.repos.d/kubernetes.repo << EOF
[kubernetes]
name=Kubernetes
baseurl=https://mirrors.tuna.tsinghua.edu.cn/kubernetes/yum/repos/kubernetes-el7-x86_64/
enabled=1
gpgcheck=0
repo_gpgcheck=0
gpgkey=https://mirrors.tuna.tsinghua.edu.cn/kubernetes/yum/doc/yum-key.gpg

EOF
```
2.5 安装kubeadm、kubelet、kubectl：
```bash
 yum install -y kubelet-1.23.0 kubeadm-1.23.0 kubectl-1.23.0
systemctl enable kubelet
```
3.1 部署k8s-master【master执行】
```bash
kubeadm init \
  --apiserver-advertise-address=X.X.X.X \
  --image-repository registry.aliyuncs.com/google_containers \
  --kubernetes-version v1.23.0 \
  --service-cidr=10.96.0.0/12 \
  --pod-network-cidr=10.244.0.0/16 \
  --ignore-preflight-errors=all
```
apiserver-advertise-address 集群通告地址
image-repository 由于默认拉取镜像地址k8s.gcr.io国内无法访问，这里指定阿里云镜像仓库地址
kubernetes-version K8s版本，与上面安装的一致
service-cidr 集群内部虚拟网络，Pod统一访问入口
pod-network-cidr Pod网络，与下面部署的CNI网络组件yaml中保持一致
初始化之后，会输出一个join命令，先复制出来，node节点加入master会使用。

默认token有效期为24小时，当过期之后，该token就不可用了。这时就需要重新创建token，可以直接使用命令快捷生成：
```bash
kubeadm token create --print-join-command
```
3.2 设置 master 节点允许调度, 默认不允许调度 pod
原因：当部署单机版的 k8s 时，这个时候 master 节点是默认不允许调度 pod 。导致创建的pod一直处于pending状态 执行以下命令, 将 master 标记为可调度即可
```bash
kubectl taint nodes --all node-role.kubernetes.io/master-
```
3.3 配置Kubectl命令自动补全功能
在使用bash命令行时，在提示符下输入某个命令的前面几个字符，然后按TAB键，就会列出以这几个字符开头的命令供我们选择。不光如此，还可以进行参数补全，但只限于文件参数，当输入到参数部分时，按TAB键，就会列出以这个参数开头的文件路径供我们选择。
依次执行以下命令进行安装

安装 bash-completion
```bash
yum -y install bash-completion
```
执行 source 命令：
```bash
source /usr/share/bash-completion/bash_completion
source <(kubectl completion bash)
echo "source <(kubectl completion bash)" >> ~/.bashrc
```
4.1 创建NFS Server


安装nfs-server
```bash
yum install -y nfs-utils
```
授权存储目录（master）
```bash
echo "/nfs/data/ *(insecure,rw,sync,no_root_squash)" > /etc/exports
```
执行以下命令，启动 nfs 服务;创建共享目录
```bash
mkdir -p /nfs/data

systemctl enable rpcbind
systemctl enable nfs-server
systemctl start rpcbind
systemctl start nfs-server
```
使配置生效
```bash
exportfs -r
```
检查配置是否生效
```bash
exportfs
```
测试，在子节点 IP为master的ip
```bash
showmount -e X.X.X.X
```
4.2 创建Service Account

管控NFS Provisioner 在k8s集群中运行的权限

*rbac.yaml: 唯一需要修改的地方只有namespace,根据实际情况定义*
```bash
yaml
--- 
apiVersion: v1
kind: ServiceAccount                 #创建个SA账号主要用来管理NFS provisioner在k8s集群中运行的权限
metadata:
  name: nfs-client-provisioner        #和上面的SA账号保持一致
  # replace with namespace where provisioner is deployed
  namespace: default
---
#以下就是ClusterRole，ClusterRoleBinding，Role，RoleBinding都是权限绑定配置，不在解释。直接复制即可。
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: nfs-client-provisioner-runner
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "update", "patch"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: run-nfs-client-provisioner
subjects:
  - kind: ServiceAccount
    name: nfs-client-provisioner
    # replace with namespace where provisioner is deployed
    namespace: default
roleRef:
  kind: ClusterRole
  name: nfs-client-provisioner-runner
  apiGroup: rbac.authorization.k8s.io
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: leader-locking-nfs-client-provisioner
  # replace with namespace where provisioner is deployed
  namespace: default
rules:
  - apiGroups: [""]
    resources: ["endpoints"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: leader-locking-nfs-client-provisioner
  # replace with namespace where provisioner is deployed
  namespace: default
subjects:
  - kind: ServiceAccount
    name: nfs-client-provisioner
    # replace with namespace where provisioner is deployed
    namespace: default
roleRef:
  kind: Role
  name: leader-locking-nfs-client-provisioner
  apiGroup: rbac.authorization.k8s.io
```

4.3 创建StorageClass 并指定 NFS provisioner

*负责建立PVC并调用NFS provisioner进行预定的工作，并让PV与PVC建立关联*
sc.yaml 需要把IP指定自己nfs服务器地址
```bash
yaml
## 创建NFS资源的StorageClass
apiVersion: storage.k8s.io/v1
#存储类的资源名称
kind: StorageClass
metadata:
 #存储类的名称，自定义
  name: nfs-storage                
  annotations:
  #注解，是否是默认的存储，注意：KubeSphere默认就需要个默认存储，因此这里注解要设置为“默认”的存储系统，表示为"true"，代表默认。
    storageclass.kubernetes.io/is-default-class: "true"          
#存储分配器的名字，自定义
provisioner: k8s-sigs.io/nfs-subdir-external-provisioner         
parameters:
  archiveOnDelete: "true"  ## 删除pv的时候，pv的内容是否要备份

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nfs-client-provisioner
  labels:
    app: nfs-client-provisioner
  # replace with namespace where provisioner is deployed
  namespace: default
spec:
  #只运行一个副本应用
  replicas: 1                 
  #描述了如何用新的POD替换现有的POD
  strategy:                   
    #Recreate表示重新创建Pod
    type: Recreate 
  #选择后端Pod           
  selector:        
    matchLabels:
      app: nfs-client-provisioner
  template:
    metadata:
      labels:
        app: nfs-client-provisioner
    spec:
      serviceAccountName: nfs-client-provisioner          #创建账户
      containers:
        - name: nfs-client-provisioner         
          image: registry.cn-hangzhou.aliyuncs.com/lfy_k8s_images/nfs-subdir-external-provisioner:v4.0.2      #使用NFS存储分配器的镜像
          # resources:
          #    limits:
          #      cpu: 10m
          #    requests:
          #      cpu: 10m
          volumeMounts:
            - name: nfs-client-root           #定义个存储卷，
              mountPath: /persistentvolumes   #表示挂载容器内部的路径
          env:
            - name: PROVISIONER_NAME          #定义存储分配器的名称
              value: k8s-sigs.io/nfs-subdir-external-provisioner         #需要和上面定义的保持名称一致
            - name: NFS_SERVER                                       #指定NFS服务器的地址，你需要改成你的NFS服务器的IP地址
              value: 192.168.11.24   ## 指定自己nfs服务器地址
            - name: NFS_PATH                                
              value: /nfs/data  ## nfs服务器共享的目录            #指定NFS服务器共享的目录
      volumes:
        - name: nfs-client-root           #存储卷的名称，和前面定义的保持一致
          nfs:
            server: 192.168.11.24            #NFS服务器的地址，和上面保持一致，这里需要改为你的IP地址
            path: /nfs/data               #NFS共享的存储目录，和上面保持一致
```
4.5 保存上面文件到/usr/local/目录下
```bash
cd /usr/local/
sudo kubectl apply -f rbac.yaml
sudo kubectl apply -f sc.yaml
```



