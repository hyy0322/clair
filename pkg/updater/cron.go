package updater

import (
	"github.com/coreos/clair"
	"github.com/coreos/clair/database"
	"github.com/coreos/clair/pkg/config"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

const UpdaterScheduleSetting = "updater-schedule"

var c *cron.Cron

func ScheduleUpdater(db database.Datastore, defaultCron string) {
	updaterSchedule, err := db.GetSetting(UpdaterScheduleSetting)
	if err != nil {
		log.Fatalf("Get setting '%s' error: %v", UpdaterScheduleSetting, err)
	}

	if len(updaterSchedule) == 0 {
		updaterSchedule = defaultCron
	}

	_, err = cron.Parse(updaterSchedule)
	if err != nil {
		log.Fatalf("Invalid cron spec '%s': %v", updaterSchedule, err)
	}

	c = cron.New()
	c.AddFunc(updaterSchedule, func() {
		log.Info("Start vulnerabilities database update by cron schedule.")
		updated := clair.Update(db)
		if updated {
			log.Infof("Vulnerabilities database updated by cron successfully.")
		} else {
			log.Errorf("Vulnerabilities database update by cron failed.")
		}
	})
	c.Start()
}

func UpdateSchedule(db database.Datastore, schedule string) error {
	_, err := cron.Parse(schedule)
	if err != nil {
		log.Errorf("Invalid cron spec '%s': %v", schedule, err)
		return err
	}

	_, err = db.UpsertSetting(UpdaterScheduleSetting, schedule)
	if err != nil {
		return err
	}

	if config.AppConfig.Updater.Disabled {
		log.Info("Updater not scheduled as disabled")
		return nil
	}

	if c != nil {
		c.Stop()
	}
	c = cron.New()
	c.AddFunc(schedule, func() {
		log.Info("Start vulnerabilities database update by cron schedule.")
		updated := clair.Update(db)
		if updated {
			log.Infof("Vulnerabilities database updated by cron successfully.")
		} else {
			log.Errorf("Vulnerabilities database update by cron failed.")
		}
	})
	c.Start()

	return nil
}

func GetSchedule(db database.Datastore) (string, error) {
	return db.GetSetting(UpdaterScheduleSetting)
}