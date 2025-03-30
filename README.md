## go_blog 

gin + vue3 开发的个人博客项目

本地测试
开发工具及版本

golang: 1.23.1

node: v22.14.0

docker: 26.1.4

编译器：vscode



### 单节点K8S部署

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

4.6 安装calico


5.1 制作前后端镜像
创建目录
```bash
mkdir -p /opt/{backend,frontend,k8s}
```

后端文件传入到/opt/backend
```bash
cd /opt/backend
```

```bash
vim Dockerfile

# 使用官方 Go 镜像作为基础镜像
FROM golang:1.23-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

# 设置工作目录
WORKDIR /app

# 将当前目录复制到容器中的 /app
COPY . .

ENV GOOS=linux
ENV GOARCH=amd64
# 编译 Go 项目
RUN go build -o blog-backend

# 使用更轻量级的镜像运行应用
FROM alpine:latest

# 安装必需的依赖，确保能运行 Go 程序
RUN apk --no-cache add ca-certificates

# 创建工作目录
WORKDIR /app

# 复制编译后的二进制文件和配置文件到容器中
COPY --from=builder /app/blog-backend /app/blog-backend
COPY config.yaml /app/config.yaml

# 设置容器的启动命令
CMD ["./blog-backend"]

# 公开后端应用的端口
EXPOSE 8080
```
```bash
docker build -t blog-backend:v1 .
```
前端文件打包dist传入到/opt/frontend
```bash
cd /opt/frontend
```
```bash
vim Dockerfile

FROM nginx:alpine
COPY dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

```bash
vim nginx.conf

server {
        listen       80;
        server_name  localhost;
	root "/usr/share/nginx/html";
	index index.html;
        location / {
        	try_files $uri $uri/ /index.html;
    }
	location /api/ {
        proxy_pass http://backend-service:8080;  # 后端地址
    }
}
```

5.2 编写k8s yaml文件
```bash
cd /opt/k8s
```

```bash
vim backend-deployment.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: blog-config
data:
  config.yaml: |
    captcha:
      height: 80
      width: 240
      length: 6
      max_skew: 0.7
      dot_count: 80
    email:
      host: smtp.qq.com
      port: 465
      from: xxx@qq.com
      nickname: 发件人昵称
      secret: xxxxxxxxxxxxxxxx
      is_ssl: true
    es:
      url: http://127.0.0.1:9200
      username: ""
      password: ""
      is_console_print: true
    gaode:
      enable: false
      key: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    jwt:
      access_token_secret: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
      refresh_token_secret: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
      access_token_expiry_time: 2h
      refresh_token_expiry_time: 7d
      issuer: go_blog
    mysql:
      host: 127.0.0.1
      port: 3306
      config: charset=utf8mb4&parseTime=True&loc=Local
      db_name: blog_db
      username: root
      password: root
      max_idle_conns: 10
      max_open_conns: 100
      log_mode: info
    qiniu:
      zone: z0
      bucket: xxx
      img_path: https://xxx.xxx/
      access_key: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
      secret_key: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
      use_https: true
      use_cdn_domains: true
    qq:
      enable: false
      app_id: "000000000"
      app_key: xxxxxxxxxxxxxxxx
      redirect_uri: http://xxx.xxx/xxx
    redis:
      address: 127.0.0.1:6379
      password: ""
      db: 0
    system:
      host: 0.0.0.0
      port: 8080
      env: release
      router_prefix: api
      use_multipoint: true
      sessions_secret: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
      oss_type: local
    upload:
      size: 20
      path: uploads
    website:
      logo: ""
      full_logo: ""
      title: 博客标题-xxx的个人博客
      slogan: 博客标题
      slogan_en: Blog Title
      description: 博客描述
      version: 1.0.0
      created_at: "2025-1-15"
      icp_filing: icp备案号
      public_security_filing: 公安备案号
      bilibili_url: https://space.bilibili.com/xxx
      gitee_url: https://gitee.com/xxx
      github_url: https://github.com/xxx/
      name: 昵称
      job: 工作
      address: 地址
      email: xxx@qq.com
      qq_image: ""
      wechat_image: ""
    zap:
      level: info
      filename: log/go_blog.log
      max_size: 200
      max_backups: 30
      max_age: 5
      is_console_print: true
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: blog-backend:v1
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: 100m    # 最少 0.1 核
            memory: 128Mi # 最少 128MB
          limits:
            cpu: 300m    # 最多 0.3 核
            memory: 512Mi # 最多 512MB
        volumeMounts:
        - name: config-volume
          mountPath: /app/config.yaml
          subPath: config.yaml
      volumes:
      - name: config-volume
        configMap:
          name: blog-config
---
apiVersion: v1
kind: Service
metadata:
  name: backend-service
spec:
  selector:
    app: backend
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
```

```bash
vim frontend-deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: frontend
        image: blog-frontend:v1
        ports:
        - containerPort: 80
        env:
        - name: VITE_SERVER_URL  # 注入后端 API 地址
          value: "http://backend-service:8080"
        #resources:
        #  requests:
        #    cpu: 50m
        #    memory: 64Mi
        #  limits:
        #    cpu: 150m
        #    memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
spec:
  selector:
    app: frontend
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
```

```bash
vim mysql-pvc.yaml

apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-pv-claim
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: nfs-storage
```

```bash
vim mysql-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mysql
  template:
    metadata:
      labels:
        app: mysql
    spec:
      containers:
      - name: mysql
        image: docker.io/library/mysql:8.0
        imagePullPolicy: IfNotPresent
        env:
          - name: MYSQL_ROOT_PASSWORD
            value: "PASSWORD"
        ports:
          - containerPort: 3306
        resources:
          requests:
            cpu: 140m
            memory: 512Mi
          limits:
            cpu: 440m
            memory: 512Mi
        volumeMounts:
          - name: mysql-pv-storage
            mountPath: /var/lib/mysql
      volumes:
      - name: mysql-pv-storage
        persistentVolumeClaim:
          claimName: mysql-pv-claim
```
```bash
vim mysql-service.yaml

apiVersion: v1
kind: Service
metadata:
  name: mysql-svc
spec:
  selector:
    app: mysql
  ports:
    - port: 3306
      targetPort: 3306
```

```bash
vim redis-pvc.yaml

apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-pv-claim
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 2Gi
  storageClassName: nfs-storage
```

```bash
vim redis-deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: docker.io/library/redis:6.0
        imagePullPolicy: IfNotPresent
        ports:
          - containerPort: 6379
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 150m
            memory: 128Mi
        volumeMounts:
          - name: redis-pv-storage
            mountPath: /data
      volumes:
      - name: redis-pv-storage
        persistentVolumeClaim:
          claimName: redis-pv-claim
```

```bash
vim redis-service.yaml

apiVersion: v1
kind: Service
metadata:
  name: redis-svc
spec:
  selector:
    app: redis
  ports:
    - port: 6379
      targetPort: 6379
```

```bash
vim es-pvc.yaml

apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: es-pv-claim
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: nfs-storage
```

```bash
vim elasticsearch-deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: elasticsearch
spec:
  replicas: 1
  selector:
    matchLabels:
      app: elasticsearch
  template:
    metadata:
      labels:
        app: elasticsearch
    spec:
      containers:
      - name: elasticsearch
        image: docker.elastic.co/elasticsearch/elasticsearch:7.10.0
        imagePullPolicy: IfNotPresent
        env:
          - name: discovery.type
            value: "single-node"  # 单节点模式
          # 可根据需求配置 JVM 参数（如 ES_JAVA_OPTS）
          - name: ES_JAVA_OPTS
            value: "-Xms512m -Xmx512m" # 限制 ES 最大 512MB 堆
        ports:
          - containerPort: 9200
          - containerPort: 9300
        resources:
          limits:
            memory: "1Gi"
            cpu: "1"
          requests:
            memory: "500Mi"
            cpu: "500m"
        volumeMounts:
          - name: es-pv-storage
            mountPath: /usr/share/elasticsearch/data
      volumes:
      - name: es-pv-storage
        persistentVolumeClaim:
          claimName: es-pv-claim
```

```bash
vim elasticsearch-service.yaml

apiVersion: v1
kind: Service
metadata:
  name: elasticsearch-svc
spec:
  selector:
    app: elasticsearch
  ports:
    - port: 9200
      targetPort: 9200
```

```bash
vim rbac.yaml

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

```bash
vim ingress-nginx.yaml

apiVersion: v1
kind: Namespace
metadata:
  labels:
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
  name: ingress-nginx
---
apiVersion: v1
automountServiceAccountToken: true
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx
  namespace: ingress-nginx
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/component: admission-webhook
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-admission
  namespace: ingress-nginx
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx
  namespace: ingress-nginx
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - configmaps
  - pods
  - secrets
  - endpoints
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses/status
  verbs:
  - update
- apiGroups:
  - networking.k8s.io
  resources:
  - ingressclasses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - coordination.k8s.io
  resourceNames:
  - ingress-nginx-leader
  resources:
  - leases
  verbs:
  - get
  - update
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - list
  - watch
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app.kubernetes.io/component: admission-webhook
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-admission
  namespace: ingress-nginx
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - create
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - endpoints
  - nodes
  - pods
  - secrets
  - namespaces
  verbs:
  - list
  - watch
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses/status
  verbs:
  - update
- apiGroups:
  - networking.k8s.io
  resources:
  - ingressclasses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - list
  - watch
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app.kubernetes.io/component: admission-webhook
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-admission
rules:
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - validatingwebhookconfigurations
  verbs:
  - get
  - update
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx
  namespace: ingress-nginx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ingress-nginx
subjects:
- kind: ServiceAccount
  name: ingress-nginx
  namespace: ingress-nginx
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app.kubernetes.io/component: admission-webhook
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-admission
  namespace: ingress-nginx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ingress-nginx-admission
subjects:
- kind: ServiceAccount
  name: ingress-nginx-admission
  namespace: ingress-nginx
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ingress-nginx
subjects:
- kind: ServiceAccount
  name: ingress-nginx
  namespace: ingress-nginx
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app.kubernetes.io/component: admission-webhook
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-admission
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ingress-nginx-admission
subjects:
- kind: ServiceAccount
  name: ingress-nginx-admission
  namespace: ingress-nginx
---
apiVersion: v1
data:
  allow-snippet-annotations: "false"
kind: ConfigMap
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-controller
  namespace: ingress-nginx
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-controller
  namespace: ingress-nginx
spec:
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - appProtocol: http
    name: http
    port: 80
    protocol: TCP
    targetPort: http
  - appProtocol: https
    name: https
    port: 443
    protocol: TCP
    targetPort: https
  selector:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
  type: NodePort
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-controller-admission
  namespace: ingress-nginx
spec:
  ports:
  - appProtocol: https
    name: https-webhook
    port: 443
    targetPort: webhook
  selector:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-controller
  namespace: ingress-nginx
spec:
  minReadySeconds: 0
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app.kubernetes.io/component: controller
      app.kubernetes.io/instance: ingress-nginx
      app.kubernetes.io/name: ingress-nginx
  strategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        app.kubernetes.io/component: controller
        app.kubernetes.io/instance: ingress-nginx
        app.kubernetes.io/name: ingress-nginx
        app.kubernetes.io/part-of: ingress-nginx
        app.kubernetes.io/version: 1.10.0
    spec:
      hostNetwork: true
      containers:
      - args:
        - /nginx-ingress-controller
        - --election-id=ingress-nginx-leader
        - --controller-class=k8s.io/ingress-nginx
        - --ingress-class=nginx
        - --configmap=$(POD_NAMESPACE)/ingress-nginx-controller
        - --validating-webhook=:8443
        - --validating-webhook-certificate=/usr/local/certificates/cert
        - --validating-webhook-key=/usr/local/certificates/key
        - --enable-metrics=false
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: LD_PRELOAD
          value: /usr/local/lib/libmimalloc.so
        image: registry.k8s.io/ingress-nginx/controller:v1.10.0
        imagePullPolicy: IfNotPresent
        lifecycle:
          preStop:
            exec:
              command:
              - /wait-shutdown
        livenessProbe:
          failureThreshold: 5
          httpGet:
            path: /healthz
            port: 10254
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        name: controller
        ports:
        - containerPort: 80
          name: http
          protocol: TCP
        - containerPort: 443
          name: https
          protocol: TCP
        - containerPort: 8443
          name: webhook
          protocol: TCP
        readinessProbe:
          failureThreshold: 3
          httpGet:
            path: /healthz
            port: 10254
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        resources:
          requests:
            cpu: 100m
            memory: 90Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - ALL
          readOnlyRootFilesystem: false
          runAsNonRoot: true
          runAsUser: 101
          seccompProfile:
            type: RuntimeDefault
        volumeMounts:
        - mountPath: /usr/local/certificates/
          name: webhook-cert
          readOnly: true
      dnsPolicy: ClusterFirst
      nodeSelector:
        kubernetes.io/os: linux
      serviceAccountName: ingress-nginx
      terminationGracePeriodSeconds: 300
      volumes:
      - name: webhook-cert
        secret:
          secretName: ingress-nginx-admission
---
apiVersion: batch/v1
kind: Job
metadata:
  labels:
    app.kubernetes.io/component: admission-webhook
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-admission-create
  namespace: ingress-nginx
spec:
  template:
    metadata:
      labels:
        app.kubernetes.io/component: admission-webhook
        app.kubernetes.io/instance: ingress-nginx
        app.kubernetes.io/name: ingress-nginx
        app.kubernetes.io/part-of: ingress-nginx
        app.kubernetes.io/version: 1.10.0
      name: ingress-nginx-admission-create
    spec:
      containers:
      - args:
        - create
        - --host=ingress-nginx-controller-admission,ingress-nginx-controller-admission.$(POD_NAMESPACE).svc
        - --namespace=$(POD_NAMESPACE)
        - --secret-name=ingress-nginx-admission
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: registry.k8s.io/ingress-nginx/kube-webhook-certgen:v1.4.0
        imagePullPolicy: IfNotPresent
        name: create
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 65532
          seccompProfile:
            type: RuntimeDefault
      nodeSelector:
        kubernetes.io/os: linux
      restartPolicy: OnFailure
      serviceAccountName: ingress-nginx-admission
---
apiVersion: batch/v1
kind: Job
metadata:
  labels:
    app.kubernetes.io/component: admission-webhook
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-admission-patch
  namespace: ingress-nginx
spec:
  template:
    metadata:
      labels:
        app.kubernetes.io/component: admission-webhook
        app.kubernetes.io/instance: ingress-nginx
        app.kubernetes.io/name: ingress-nginx
        app.kubernetes.io/part-of: ingress-nginx
        app.kubernetes.io/version: 1.10.0
      name: ingress-nginx-admission-patch
    spec:
      containers:
      - args:
        - patch
        - --webhook-name=ingress-nginx-admission
        - --namespace=$(POD_NAMESPACE)
        - --patch-mutating=false
        - --secret-name=ingress-nginx-admission
        - --patch-failure-policy=Fail
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: registry.k8s.io/ingress-nginx/kube-webhook-certgen:v1.4.0
        imagePullPolicy: IfNotPresent
        name: patch
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 65532
          seccompProfile:
            type: RuntimeDefault
      nodeSelector:
        kubernetes.io/os: linux
      restartPolicy: OnFailure
      serviceAccountName: ingress-nginx-admission
---
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: nginx
spec:
  controller: k8s.io/ingress-nginx
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  labels:
    app.kubernetes.io/component: admission-webhook
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
    app.kubernetes.io/version: 1.10.0
  name: ingress-nginx-admission
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    service:
      name: ingress-nginx-controller-admission
      namespace: ingress-nginx
      path: /networking/v1/ingresses
  failurePolicy: Fail
  matchPolicy: Equivalent
  name: validate.nginx.ingress.kubernetes.io
  rules:
  - apiGroups:
    - networking.k8s.io
    apiVersions:
    - v1
    operations:
    - CREATE
    - UPDATE
    resources:
    - ingresses
  sideEffects: None
```

```bash
vim my-ingress.yaml

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
  namespace: default
spec:
  ingressClassName: nginx
  rules:
    - host: www.xlybd.com
      http:
        paths:
          - path: /api/
            pathType: Prefix
            backend:
              service:
                name: backend-service
                port:
                  number: 8080
          - path: /
            pathType: Prefix
            backend:
              service:
                name: frontend-service
                port:
                  number: 80
```

```bash
kubectl apply -f ingress-nginx.yaml
kubectl apply -f rbac.yaml
kubectl apply -f mysql-pvc.yaml
kubectl apply -f mysql-deployment.yaml
kubectl apply -f mysql-service.yaml
kubectl apply -f redis-pvc.yaml
kubectl apply -f redis-deployment.yaml
kubectl apply -f redis-service.yaml
kubectl apply -f es-pvc.yaml
kubectl apply -f elasticsearch-deployment.yaml
kubectl apply -f elasticsearch-service.yaml
kubectl apply -f backend-deployment.yaml
kubectl apply -f frontend-deployment.yaml
kubectl apply -f my-ingress.yaml
```
