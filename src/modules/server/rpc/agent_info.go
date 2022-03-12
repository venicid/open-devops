package rpc

import (
	"encoding/json"
	"fmt"
	"log"
	"open-devops/src/models"
)

func (r *Server) HostInfoReport(input models.AgentCollectInfo, output *string)  error{
	log.Printf("[HostInfoReport][input:%+v]", input)

	// 统一字段
	ips := []string{input.IpAddr}
	ipJ,_ := json.Marshal(ips)
	if input.SN == ""{
		input.SN = input.HostName
	}
	if input.SN == ""{
		*output = "sn.empty"
		return nil
	}

	// 先获取对象的uid
	rh := models.ResourceHost{
		Uid:               input.SN,
		Hash:              "",
		Name:              input.HostName,
		PrivateIps:        ipJ,
		CPU:               input.CPU,
		Mem:               input.Mem,
		Disk:              input.Disk,
	}
	hash := rh.GetHash()

	// 用uid去db中获取之前的结果，再根据两者之间的hash是否一致，决定要改
	rhUid := models.ResourceHost{Uid: input.SN}
	rhUidDb, err := rhUid.GetOne()
	if err != nil{
		*output = fmt.Sprintf("db_getone_err_%v", err)
		return err
	}


	if rhUidDb == nil{
		// 说明指定uid不存在，插入
		rh.Hash = hash
		err = rh.AddOne()

		if err != nil{
			*output = fmt.Sprintf("db_err_%v", err)
			return err
		}else{
			*output = "insert_success"
		}
		return nil
	}

	// uid存在需要判断hash
	if rhUidDb.Hash != hash{
		rh.Hash = hash
		updated, err := rh.Update()
		if err!= nil{
			*output = fmt.Sprintf("update_error_%v", err)
			return err
		}
		if updated{
			*output = "update_success"
			return nil
		}
	}

	// uid存在且hash值相等，什么都不需要做
	log.Printf("[host.info.same][input:%+v]", input)

	return nil
}