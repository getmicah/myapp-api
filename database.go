package main

import (
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
)

// GetStationStatus : get userID value
func GetStationStatus(db *bolt.DB, userID string) (bool, error) {
	var status bool
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(StationsBucket))
		v := b.Get([]byte(userID))
		if string(v[:]) == StationActive {
			status = true
		} else {
			status = false
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return status, nil
}

// TurnOnStation : set userID to "on"
func TurnOnStation(db *bolt.DB, userID string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(StationsBucket))
		v := b.Get([]byte(userID))
		if string(v[:]) == StationActive {
			msg := fmt.Sprintf("%s station is already on", userID)
			return errors.New(msg)
		}
		err := b.Put([]byte(userID), []byte(StationActive))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// TurnOffStation : set userID to "off"
func TurnOffStation(db *bolt.DB, userID string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(StationsBucket))
		v := b.Get([]byte(userID))
		if string(v[:]) != StationActive {
			msg := fmt.Sprintf("%s station is already off", userID)
			return errors.New(msg)
		}
		err := b.Put([]byte(userID), []byte(""))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
