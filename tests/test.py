
import curlify
import requests

JSON_H = {'Content-Type': 'application/json'}


def node_path_add():
    data = {
        "node": "a1.b1.c1"

    }
    print(data)
    uri = 'http://localhost:8082/api/v1/node-path'
    res = requests.post(uri, json=data)
    print(res.status_code)
    print(res.text)


def node_path_query():
    data = {
        "node": "a1",
        "query_type":2,

    }
    print(data)
    uri = 'http://localhost:8082/api/v1/node-path'
    res = requests.get(uri, json=data, headers=JSON_H)
    print(curlify.to_curl(res.request))
    print(res.status_code)
    print(res.text)

def resource_mount():
    data = {
        "target_path": "waimai.ditu.es",
        "resource_type":"resource_host",
        "resource_ids":[1],

    }
    print(data)
    uri = 'http://localhost:8082/api/v1/resource-mount'
    res = requests.post(uri, json=data, headers=JSON_H)
    print(curlify.to_curl(res.request))
    print(res.status_code)
    print(res.text)

def resource_unmount():
    data = {
        "target_path": "waimai.ditu.es",
        "resource_type":"resource_host",
        "resource_ids":[1],

    }
    print(data)
    uri = 'http://localhost:8082/api/v1/resource-unmount'
    res = requests.delete(uri, json=data, headers=JSON_H)
    print(curlify.to_curl(res.request))
    print(res.status_code)
    print(res.text)

def resource_query():

    matcher1 = {
               "key": "stree_app",
               "value": "zookeeper",
               "type": 1,
               }

    matcher12 = {
               "key": "stree_app",
               "value": "kafaka|prometheus",
               "type": 3,
               }

    matcher13 = {
               "key": "stree_group",
               "value": "inf",
               "type": 1,
               }

    matcher2 = {
               "key": "name",
               "value": "genMockResourceHost_host_3",
               "type": 1,
               }
    matcher3 =  {
            "key": "private_ips",
            "value": "8.*.8.*",
            "type": 3,
        }

    matcher4 =  {
            "key": "os",
            "value": "amd64",
            "type": 2,   # 类型 1-4 = != ~= ~!
        }

    test1 =  {
            "key": "region",
            "value": "beijing",
            "type": 1,
        }

    data = {
        "resource_type":"resource_host",
        "labels":[
#             matcher1,
#             matcher2,
            matcher12,
            matcher13,
        ],
        "target_label": "cluster"
    }
    print(data)
    g_params = {
    "page_size":2000
    }

    uri = 'http://localhost:8082/api/v1/resource-query'
    res = requests.post(uri, json=data, params=g_params,  headers=JSON_H)
    print(curlify.to_curl(res.request))
    print(res.status_code )
    print(res.text)
    data = res.json().get("result")
    data = data if data is not None else []
    print(len(data))
    for i in data:
        print(i)

# node_path_add()
# node_path_query()
# resource_mount()
# resource_unmount()
resource_query()

"""
测试倒排索引resource_query，对比结果
关闭sync同步，配置文件
    public_cloud_sync:
      enable: false

运行脚本
    python test3.py

验证数量是否一致
    - 在sql中执行 select * from resource_host rh  where region="beijing";


"""