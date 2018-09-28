// Copyright (C) 2017 Battelle Memorial Institute
// Copyright (C) 2018 Chunguang Wu <chokko@126.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform/helper/schema"
	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

func resourceGlusterVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceGlusterVolumeCreate,
		Read:   resourceGlusterVolumeRead,
		Delete: resourceGlusterVolumeDelete,
		Update: resourceGlusterVolumeUpdate,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The cluster id of the gluster volume belong to",
			},
			"volume_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"replica_count": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},
			"bricks": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: false,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server_id": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "The server id of the brick in",
						},
						"brick_dir": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "Brick dir ",
						},
						"brick_id": {
							Type:        schema.TypeString,
							Required:    false,
							ForceNew:    false,
							Computed:    true,
							Description: "Brick id",
						},
						"brick_name": {
							Type:        schema.TypeString,
							Required:    false,
							ForceNew:    false,
							Computed:    true,
							Description: "Brick name",
						},
					},
				},
			},
			"volume_status": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "up",
				Description: "Set the volume to be started of stopped",
			},
			"volume_rebalance": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Set the volume should rebalance or not, nil for do nothing, start for start rebalance,stop for stop rebalance.",
			},
		},
	}
}

func resourceGlusterVolumeCreate(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*ovirtsdk4.Connection)

	volumeName := d.Get("name").(string)
	clusterID := d.Get("cluster_id").(string)
	volumeTypeOri := d.Get("volume_type").(string)
	replicaCount := d.Get("replica_count").(int)
	volumeStatus := d.Get("volume_status").(string)
	//volumeRebalance := d.Get("volume_rebalance").(string)
	var replicaCount64 int64
	var volumeType ovirtsdk4.GlusterVolumeType
	volumeType = ovirtsdk4.GlusterVolumeType(volumeTypeOri)
	var bricks []interface{}
	if b, ok := d.GetOk("bricks"); ok {
		bricks = b.([]interface{})
	}
	//validate input
	if volumeType == "distributed_replicate" {
		if len(bricks)%replicaCount != 0 {
			return fmt.Errorf("Brick num mismatch relicate number")
		}
	}
	if volumeType == "replicate" {
		if len(bricks) != replicaCount {
			return fmt.Errorf("Number of bricks must be a equal to replica count for a REPLICATE volume")
		}
	}
	if volumeType == "distribute" {
		if replicaCount != 0 {
			return fmt.Errorf("Distribute volume  replica count should be 0")
		}
	}

	clustersService := conn.SystemService().ClustersService()
	cluster := clustersService.ClusterService(clusterID)

	glusterBricks := make([]*ovirtsdk4.GlusterBrick, len(bricks))
	for _, v := range bricks {
		vmap := v.(map[string]interface{})
		brickOne := new(ovirtsdk4.GlusterBrick)
		brickOne.SetBrickDir(vmap["brick_dir"].(string))
		brickOne.SetServerId(vmap["server_id"].(string))
		glusterBricks = append(glusterBricks, brickOne)
	}

	//construct GlusterBrickSlice struct
	glusterBricksSliceNew := new(ovirtsdk4.GlusterBrickSlice)
	glusterBricksSliceNew.SetSlice(glusterBricks)
	//construct GlusterVolume
	glusterVolumeNew := new(ovirtsdk4.GlusterVolume)
	glusterVolumeNew.SetName(volumeName)
	glusterVolumeNew.SetVolumeType(volumeType)
	glusterVolumeNew.SetBricks(glusterBricksSliceNew)
	if volumeType == "distributed_replicate" || volumeType == "replicate" {

		replicaCount64 = int64(replicaCount)
		glusterVolumeNew.SetReplicaCount(replicaCount64)
	} else if volumeType == "distribute" {
		if replicaCount != 0 {
			return fmt.Errorf("Distributed volume replica count should be 0")
		}
	} else {
		return fmt.Errorf("Only support distribute,replicate and distributed_replicate volume")
	}

	volumeResponse, err := cluster.GlusterVolumesService().Add().Volume(glusterVolumeNew).Query("force", "true").Send()
	if err != nil {
		return fmt.Errorf("Add volume failed, reason: %s", err)
	}
	volume := volumeResponse.MustVolume()
	volumeStatusNow := volume.MustStatus()
	d.SetId(volume.MustId())
	volumeService := cluster.GlusterVolumesService().VolumeService(d.Id())
	//start or stop volume
	if volumeStatus == "up" && volumeStatusNow == "down" {
		_, err = volumeService.Start().Async(false).Send()
		if err != nil {
			return fmt.Errorf("Start volume failed, reason: %s", err)
		}
	}
	if volumeStatus == "down" && volumeStatusNow == "up" {
		_, err = volumeService.Stop().Send()
		if err != nil {
			return fmt.Errorf("Stop volume failed, reason: %s", err)
		}
	}
	//rebalance volume
	//first create not need rebalance

	return resourceGlusterVolumeRead(d, meta)

}

func resourceGlusterVolumeRead(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*ovirtsdk4.Connection)
	clusterID := d.Get("cluster_id").(string)
	clustersService := conn.SystemService().ClustersService()
	cluster := clustersService.ClusterService(clusterID)

	volume := cluster.GlusterVolumesService().VolumeService(d.Id())
	volumeResponse, err := volume.Get().Send()
	if err != nil {
		if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Get volume  failed, reason: %s", err)
	}

	if volumeResponseVolume, ok := volumeResponse.Volume(); ok {

		d.Set("name", volumeResponseVolume.MustName())
		d.Set("volume_type", volumeResponseVolume.MustVolumeType())
		d.Set("replica_count", volumeResponseVolume.MustReplicaCount())
		d.Set("volume_status", volumeResponseVolume.MustStatus())

		bricks, err := volume.GlusterBricksService().List().Send()
		if err != nil {
			return fmt.Errorf("Get bricks from volume failed: %s", err)
		}

		if err = d.Set("bricks", fillBricks(bricks.MustBricks().Slice())); err != nil {
			return fmt.Errorf("Error setting bricks: %s", err)
		}

		return nil

	}

	return fmt.Errorf("Volume is nil")

}

func resourceGlusterVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	clusterID := d.Get("cluster_id").(string)
	clustersService := conn.SystemService().ClustersService()
	cluster := clustersService.ClusterService(clusterID)

	volume := cluster.GlusterVolumesService().VolumeService(d.Id())
	volumeResponse, err := volume.Get().Send()
	if err != nil {
		if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
			return nil
		}
		return fmt.Errorf("Get volume  failed, reason: %s", err)
	}
	volumeStatusNow := volumeResponse.MustVolume().MustStatus()
	if volumeStatusNow == "up" {
		_, err := volume.Stop().Send()
		if err != nil {
			return fmt.Errorf("Stop volume failed, reason: %s", err)
		}
	}
	_, err = volume.Remove().Send()
	if err != nil {
		return fmt.Errorf("Error remove gluster volume: %s", err)
	}
	return nil
}
func resourceGlusterVolumeUpdate(d *schema.ResourceData, meta interface{}) error {

	//volume name ,cluster_id disallow to change
	if d.HasChange("name") {
		return fmt.Errorf("Name change is not supported")
	}
	if d.HasChange("cluster_id") {
		return fmt.Errorf("Cluster id change is not supported")
	}

	conn := meta.(*ovirtsdk4.Connection)
	clusterID := d.Get("cluster_id").(string)
	clustersService := conn.SystemService().ClustersService()
	cluster := clustersService.ClusterService(clusterID)

	volume := cluster.GlusterVolumesService().VolumeService(d.Id())

	_, err := volume.Get().Send()
	if err != nil {
		if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Get volume  failed, reason: %s", err)
	}

	d.Partial(true)

	//volume_status change
	var oldVolumeStatus, newVolumeStatus string
	if d.HasChange("volume_status") {
		ovs, nvs := d.GetChange("volume_status")
		newVolumeStatus = nvs.(string)
		oldVolumeStatus = ovs.(string)
		if newVolumeStatus == "up" && oldVolumeStatus == "down" {
			_, err := volume.Start().Send()
			if err != nil {
				return fmt.Errorf("Start volume failed, reason: %s", err)
			}
		}

		if newVolumeStatus == "down" && oldVolumeStatus == "up" {
			_, err := volume.Stop().Send()
			if err != nil {
				return fmt.Errorf("Stop volume failed, reason: %s", err)
			}
		}

		d.SetPartial("volume_status")
	} else {
		oldVolumeStatus = d.Get("volume_status").(string)
		newVolumeStatus = d.Get("volume_status").(string)
	}

	//get replica_count
	var oldReplicaCount, newReplicaCount int
	if d.HasChange("replica_count") {
		orc, nrc := d.GetChange("replica_count")
		oldReplicaCount = orc.(int)
		newReplicaCount = nrc.(int)
		if !d.HasChange("bricks") {
			return fmt.Errorf("Bricks not changed")
		}
	} else {
		oldReplicaCount = d.Get("replica_count").(int)
		newReplicaCount = d.Get("replica_count").(int)
	}

	//get volume type,volume type change check,only distributed_replicate ->replicate(or distribute)->distribute allowed
	var oldVolumeType, newVolumeType string
	if d.HasChange("volume_type") {
		ovt, nvt := d.GetChange("volume_type")
		oldVolumeType = ovt.(string)
		newVolumeType = nvt.(string)
		if oldVolumeType == "distribute" {
			return fmt.Errorf("Volume type change error,distribute type can  not change to other type")
		}
		if oldVolumeType == "replicate" && newVolumeType == "distributed_replicate" {

			return fmt.Errorf("Volume type change error,replicate type can  not change to distributed_replicate type")
		}
		if !d.HasChange("bricks") {
			return fmt.Errorf("Bricks not changed")
		}
	} else {
		oldVolumeType = d.Get("volume_type").(string)
		newVolumeType = d.Get("volume_type").(string)
	}

	//brick change
	if d.HasChange("bricks") {
		obn, nbn := d.GetChange("bricks")
		oldBricksNum := obn.([]interface{})
		newBricksNum := nbn.([]interface{})
		oldBricksLen := len(oldBricksNum)
		newBricksLen := len(newBricksNum)
		//add bricks,replica count should not decrease,volume type should not change
		if newBricksLen > oldBricksLen {
			if newReplicaCount < oldReplicaCount {
				return fmt.Errorf("Replica count should increase")
			}
			if newVolumeType != oldVolumeType {
				return fmt.Errorf("Volume type should not change when add bricks")
			}
			if newVolumeType == "replica" || newVolumeType == "distributed_replicate" {
				if newBricksLen%newReplicaCount != 0 {
					return fmt.Errorf("Number of bricks should be a mutiple of Replica Count")
				}
			}
			//check added bricks in order
			if !addedBrickInOrder(oldBricksNum, newBricksNum) {
				return fmt.Errorf("Added bricks should appended to the  original bricks")

			}
			//add bricks action
			glusterBricks := make([]*ovirtsdk4.GlusterBrick, newBricksLen)
			for _, v := range newBricksNum[oldBricksLen:] {
				vmap := v.(map[string]interface{})
				brickOne := new(ovirtsdk4.GlusterBrick)
				brickOne.SetBrickDir(vmap["brick_dir"].(string))
				brickOne.SetServerId(vmap["server_id"].(string))
				glusterBricks = append(glusterBricks, brickOne)
			}

			//construct GlusterBrickSlice struct
			glusterBricksSliceNew := new(ovirtsdk4.GlusterBrickSlice)
			glusterBricksSliceNew.SetSlice(glusterBricks)
			//add bricks to volume
			_, err := volume.GlusterBricksService().Add().Bricks(glusterBricksSliceNew).ReplicaCount(int64(newReplicaCount)).Query("force", "true").Send()
			if err != nil {
				return fmt.Errorf("Add bricks failed, reason: %s", err)
			}

		} else if newBricksLen < oldBricksLen {
			//brick remove action
			replicateCountChange := ""
			//construct GlusterBrickSlice to remove
			glusterBricks := make([]*ovirtsdk4.GlusterBrick, oldBricksLen-newBricksLen)
			for _, v := range oldBricksNum {
				vo := v.(map[string]interface{})
				if !oneBrickIn(vo, newBricksNum) {
					brickOne := new(ovirtsdk4.GlusterBrick)
					//brickOne.SetBrickDir(vo["brick_dir"].(string))
					//brickOne.SetServerId(vo["server_id"].(string))
					brickOne.SetName(vo["brick_name"].(string))
					brickOne.SetId(vo["brick_id"].(string))
					glusterBricks = append(glusterBricks, brickOne)
				}
			}
			//construct GlusterBrickSlice struct
			glusterBricksSliceNew := new(ovirtsdk4.GlusterBrickSlice)
			glusterBricksSliceNew.SetSlice(glusterBricks)

			if newBricksLen == 0 {
				return fmt.Errorf("You should not remove all bricks")
			}
			//check left bricks in volume
			if !leftBricksIn(newBricksNum, oldBricksNum) {
				return fmt.Errorf("Left bricks for delete should in volume")
			}
			//replicate volume remove bricks
			if oldVolumeType == "replicate" {
				//delete bricks should change replica count
				if newReplicaCount != newBricksLen && newBricksLen != 1 {
					return fmt.Errorf("Replica count mismatch bricks length")
				}
				//replica->distribute
				if newBricksLen == 1 {
					if newVolumeType != "distribute" {
						return fmt.Errorf("Volume type should change to distribute")
					}
					if newReplicaCount != 0 {
						return fmt.Errorf("Distributed volume replica count should be 0")
					}
					replicateCountChange = "reptodis"
				}

				//delete bricks to volume
				if replicateCountChange == "" {
					_, err := volume.GlusterBricksService().Remove().Bricks(glusterBricksSliceNew).Send()
					if err != nil {
						return fmt.Errorf("Remove bricks failed, reason: %s", err)
					}
				}
				if replicateCountChange == "reptodis" {
					_, err := volume.GlusterBricksService().Remove().ReplicaCount(int64(oldReplicaCount - 1)).Bricks(glusterBricksSliceNew).Send()
					if err != nil {
						return fmt.Errorf("Remove bricks failed, reason: %s", err)
					}
				}

			} else if oldVolumeType == "distribute" {

				//delete bricks to volume
				_, err := volume.GlusterBricksService().Remove().Bricks(glusterBricksSliceNew).Send()
				if err != nil {
					return fmt.Errorf("Remove bricks failed, reason: %s", err)
				}

			} else if oldVolumeType == "distributed_replicate" {
				//distributed_replicate volume delete
				//remove a subvolume
				if oldBricksLen-newBricksLen == oldReplicaCount && removedBricksInSubvolume(newBricksNum, oldBricksNum, oldReplicaCount) {

					//distributed_replicate->replicate
					if newBricksLen == oldReplicaCount {
						if newVolumeType != "replicate" {
							return fmt.Errorf("Volume type should change to replicate")
						}
					}
					//delete bricks to volume
					_, err := volume.GlusterBricksService().Remove().Bricks(glusterBricksSliceNew).Send()
					if err != nil {
						return fmt.Errorf("Remove bricks failed, reason: %s", err)
					}

				} else if oldBricksLen-newBricksLen == oldBricksLen/oldReplicaCount && removedEveryBrickInSubvolume(newBricksNum, oldBricksNum, oldReplicaCount) {
					//remove one brick in every subvolume,to do remove multi bricks in every subvolume
					//check replicate count and replicate type
					//distributed_replicate->distribute
					if newBricksLen == oldBricksLen/oldReplicaCount {
						if newVolumeType != "distribute" {
							return fmt.Errorf("Volume type should change to distribute")
						}
						if newReplicaCount != 0 {
							return fmt.Errorf("Distributed volume replica count should be 0")
						}
					} else {
						if newVolumeType != oldVolumeType {
							return fmt.Errorf("Volume type should not change ")
						}
						if newReplicaCount != oldReplicaCount-1 {
							return fmt.Errorf("Replica count should decrease  1")
						}
					}
					//delete bricks action
					//delete bricks to volume
					_, err := volume.GlusterBricksService().Remove().ReplicaCount(int64(oldReplicaCount - 1)).Bricks(glusterBricksSliceNew).Send()
					if err != nil {
						return fmt.Errorf("Remove bricks failed, reason: %s", err)
					}
				} else {
					return fmt.Errorf("Removed bricks should in a subvolume or every brick in every subvolume")
				}

			}
		}
		d.SetPartial("bricks")
	}

	// If we were to return here, before disabling partial mode below,
	// then only the "address" field would be saved.

	// We succeeded, disable partial mode. This causes Terraform to save
	// all fields again.
	d.Partial(false)
	//volume_rebalance,last rebalance
	if d.HasChange("volume_rebalance") {
		_, n := d.GetChange("volume_rebalance")
		nvr := n.(string)
		if len(d.Get("bricks").([]interface{})) >= 1 && nvr == "start" && oldVolumeStatus == "up" && (oldVolumeType == "distribute" || oldVolumeType == "distributed_replicate") {
			_, err = volume.Rebalance().Async(false).Send()
			if err != nil {
				return fmt.Errorf("Rebalance volume failed, reason: %s", err)
			}

		}
		if len(d.Get("bricks").([]interface{})) >= 1 && nvr == "stop" && oldVolumeStatus == "up" && (oldVolumeType == "distribute" || oldVolumeType == "distributed_replicate") {
			_, err = volume.StopRebalance().Async(false).Send()
			if err != nil {
				return fmt.Errorf("Rebalance volume failed, reason: %s", err)
			}

		}

	}
	return resourceGlusterVolumeRead(d, meta)
}
func fillBricks(bricksSlice []*ovirtsdk4.GlusterBrick) []map[string]interface{} {
	bricks := make([]map[string]interface{}, len(bricksSlice))
	for i, v := range bricksSlice {
		attrs := make(map[string]interface{})

		if serverId, ok := v.ServerId(); ok {
			attrs["server_id"] = serverId
		}
		if brickDir, ok := v.BrickDir(); ok {
			attrs["brick_dir"] = brickDir
		}
		if brickId, ok := v.Id(); ok {
			attrs["brick_id"] = brickId
		}
		if brickName, ok := v.Name(); ok {
			attrs["brick_name"] = brickName
		}
		bricks[i] = attrs
	}
	return bricks
}
func addedBrickInOrder(oriSlice []interface{}, nowSlice []interface{}) bool {
	for i, v := range oriSlice {
		vo := v.(map[string]interface{})
		if !reflect.DeepEqual(vo, nowSlice[i].(map[string]interface{})) {
			return false
		}
	}
	return true
}
func leftBricksIn(oriSlice []interface{}, nowSlice []interface{}) bool {
	flagSlice := make([]int, len(oriSlice))
	for i, v := range oriSlice {
		vo := v.(map[string]interface{})

		for _, nv := range nowSlice {
			vn := nv.(map[string]interface{})
			if vo["server_id"] == vn["server_id"] && vo["brick_dir"] == vn["brick_dir"] {
				flagSlice[i] = 1
				break
			}
		}
	}
	for _, fv := range flagSlice {
		if fv == 0 {
			return false
		}
	}
	return true
}
func oneBrickIn(onebrick map[string]interface{}, nowSlice []interface{}) bool {
	for _, v := range nowSlice {
		vo := v.(map[string]interface{})
		if vo["server_id"] == onebrick["server_id"] && vo["brick_dir"] == onebrick["brick_dir"] {
			return true
		}

	}
	return false
}
func removedBricksInSubvolume(oriSlice []interface{}, nowSlice []interface{}, relicateCount int) bool {
	var removeSlice []int
	for i, v := range nowSlice {
		vo := v.(map[string]interface{})
		flagIn := 1
		for _, nv := range oriSlice {
			vn := nv.(map[string]interface{})
			if vo["server_id"] == vn["server_id"] && vo["brick_dir"] == vn["brick_dir"] {
				flagIn = 0
				break
			}
		}
		if flagIn == 1 {
			removeSlice = append(removeSlice, i)
		}
	}

	flagSeq := 1
	for i, _ := range removeSlice {
		if i < len(removeSlice)-1 {
			if removeSlice[i]+1 != removeSlice[i+1] {
				flagSeq = 0
				return false
			}
		}
	}
	if flagSeq == 1 {
		if (removeSlice[0])%relicateCount == 0 {
			return true
		}
	}
	return false

}

func removedEveryBrickInSubvolume(oriSlice []interface{}, nowSlice []interface{}, relicateCount int) bool {
	//removeSlice like [0,3,6] for replicateCount 3
	var removeSlice []int
	//nowSeqSlice like [0,1,2,3,4,5,6,7,8] for replicateCount 3
	var nowSeqSlice []int
	for i, v := range nowSlice {
		nowSeqSlice = append(nowSeqSlice, i)
		vo := v.(map[string]interface{})
		flagIn := 1
		for _, nv := range oriSlice {
			vn := nv.(map[string]interface{})
			if vo["server_id"] == vn["server_id"] && vo["brick_dir"] == vn["brick_dir"] {
				flagIn = 0
				break
			}
		}
		if flagIn == 1 {
			removeSlice = append(removeSlice, i)
		}
	}
	flagSlice := make([]int, len(removeSlice))
	step := 0
	for i, v := range removeSlice {
		if Contains(nowSeqSlice[step:step+relicateCount], v) {
			flagSlice[i] = 1
		}
		step = step + relicateCount
	}
	for _, fv := range flagSlice {
		if fv == 0 {
			return false
		}
	}
	return true

}
func Contains(a []int, x int) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
