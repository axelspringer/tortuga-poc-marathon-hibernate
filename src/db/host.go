package db

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	// ActionCooldownSeconds time to prevent race situations
	ActionCooldownSeconds = 2
)

// HostState enum type
type HostState int

const (
	// HostStateRun state
	HostStateRun HostState = iota
	// HostStateHibernate state
	HostStateHibernate
	// HostStateWakeUp state
	HostStateWakeUp
	// HostStateDozeOff state
	HostStateDozeOff
)

// String representation
func (hs HostState) String() string {
	return [...]string{"run", "hibernate", "wakeup", "dozeoff"}[hs]
}

// Parse string to HostState
func (hs *HostState) Parse(s string) error {
	switch s {
	case "run":
		*hs = HostStateRun
		return nil
	case "hibernate":
		*hs = HostStateHibernate
		return nil
	case "wakeup":
		*hs = HostStateWakeUp
		return nil
	case "dozeoff":
		*hs = HostStateDozeOff
		return nil
	}

	return fmt.Errorf("HostState parse: unknown type %s", s)
}

// HostEntry model
type HostEntry struct {
	ID              string         `json:"id"`
	Hosts           []string       `json:"hosts"`
	LatestUsage     int64          `json:"latestUsage"`
	ActionNotBefore int64          `json:"actionNotBefore"`
	State           string         `json:"state"`
	ScaleMap        map[string]int `json:"scaleMap"`
	IdleDuration    int64          `json:"idleDuration"`
}

// ToJSON marshal
func (he *HostEntry) ToJSON() string {
	buffer, _ := json.Marshal(he)
	return string(buffer)
}

// HostEntryManager interface
type HostEntryManager struct {
	db *Connector
}

// NewHostEntryManager creates a new instance
func NewHostEntryManager(db *Connector) *HostEntryManager {
	hem := &HostEntryManager{
		db: db,
	}

	return hem
}

// UpdateLatestUsage updates the state
func (hem *HostEntryManager) UpdateLatestUsage(groupID string) error {
	log.Infof("%s update latest usage", groupID)
	currentTS := time.Now().Unix()

	_, err := hem.db.Service.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(hem.db.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(groupID),
			},
		},
		UpdateExpression: aws.String("set latestUsage = :t"),
		ReturnValues:     aws.String("UPDATED_NEW"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":t": {
				N: aws.String(strconv.FormatInt(currentTS, 10)),
			},
		},
	})

	return err
}

// UpdateState updates the state
func (hem *HostEntryManager) UpdateState(groupID string, state HostState) error {
	currentTS := time.Now().Unix() + ActionCooldownSeconds

	log.Infof("%s update state => %s", groupID, state.String())

	_, err := hem.db.Service.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(hem.db.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(groupID),
			},
		},
		UpdateExpression: aws.String("set #st = :s, actionNotBefore = :a"),
		ExpressionAttributeNames: map[string]*string{
			"#st": aws.String("state"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":s": {
				S: aws.String(state.String()),
			},
			":a": {
				N: aws.String(strconv.FormatInt(currentTS, 10)),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	})

	return err
}

// UpdateScaleMap updates the host scale map
func (hem *HostEntryManager) UpdateScaleMap(groupID string, scaleMap map[string]int) error {
	currentTS := time.Now().Unix()
	smAttr, err := dynamodbattribute.Marshal(scaleMap)

	if err != nil {
		return err
	}

	_, err = hem.db.Service.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(hem.db.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(groupID),
			},
		},
		UpdateExpression: aws.String("set scaleMap = :c, actionNotBefore = :a"),
		ReturnValues:     aws.String("UPDATED_NEW"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":c": smAttr,
			":a": {
				N: aws.String(strconv.FormatInt(currentTS, 10)),
			},
		},
	})

	return err
}

// GetAllEntries returns a list of HostEntry
func (hem *HostEntryManager) GetAllEntries() []HostEntry {
	itemList, err := hem.db.Service.Scan(&dynamodb.ScanInput{
		TableName: aws.String(hem.db.TableName),
	})

	if err != nil {
		log.Warn(err)
		return []HostEntry{}
	}

	recs := []HostEntry{}
	err = dynamodbattribute.UnmarshalListOfMaps(itemList.Items, &recs)
	if err != nil {
		log.Warn(err)
		return []HostEntry{}
	}

	return recs
}

// GetByHost returns a HostEntry by host
func (hem *HostEntryManager) GetByHost(host string) *HostEntry {
	return nil
}

// GetByState returns a HostEntry list by state
func (hem *HostEntryManager) GetByState(state HostState) []*HostEntry {
	return nil
}
