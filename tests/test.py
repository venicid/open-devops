
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
    data = {
        "resource_type":"resource_host",
        "labels":[
        {
            "key": "group",
            "value": "sgt",
            "type": 1,
        }
        ],
        "target_label": "cluster"
    }
    print(data)
    uri = 'http://localhost:8082/api/v1/resource-query?page_size=1&page_count=1'
    res = requests.post(uri, json=data, headers=JSON_H)
    print(curlify.to_curl(res.request))
    print(res.status_code)
    print(res.text)

# node_path_add()
# node_path_query()
# resource_mount()
# resource_unmount()
resource_query()