package engine

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

const (
	// Freswitch event property names
	FS_CDR_MAP            = "variables"
	FS_DIRECTION          = "direction"
	FS_UUID               = "uuid" // -Unique ID for this call leg
	FS_CALL_DEST_NR       = "dialed_extension"
	FS_PARK_TIME          = "start_epoch"
	FS_SETUP_TIME         = "start_epoch"
	FS_ANSWER_TIME        = "answer_epoch"
	FS_HANGUP_TIME        = "end_epoch"
	FS_DURATION           = "billsec"
	FS_USERNAME           = "user_name"
	FS_IP                 = "sip_local_network_addr"
	FS_CDR_SOURCE         = "freeswitch_json"
	FS_SIP_REQUSER        = "sip_req_user" // Apps like FusionPBX do not set dialed_extension, alternative being destination_number but that comes in customer profile, not in vars
	FS_PROGRESS_MEDIAMSEC = "progress_mediamsec"
	FS_PROGRESSMS         = "progressmsec"
	FS_SIP_FROM_USER      = "sip_from_user"
	FS_SIP_TO_USER        = "sip_to_user"
)

func NewFSCdr(body []byte, cfg *config.Config) (*FSCdr, error) {
	fsCdr := &FSCdr{cfg: cfg, vars: make(map[string]string)}
	var err error
	if err = json.Unmarshal(body, &fsCdr.body); err == nil {
		if variables, ok := fsCdr.body[FS_CDR_MAP]; ok {
			if variables, ok := variables.(map[string]interface{}); ok {
				for k, v := range variables {
					fsCdr.vars[k] = v.(string)
				}
			}
			//log.Print("VARS: ", utils.ToIJSON(fsCdr.vars))
			return fsCdr, nil
		}
	}
	return nil, err
}

type FSCdr struct {
	cfg  *config.Config
	vars map[string]string
	body map[string]interface{} // keeps the loaded body for extra field search
}

func (fsCdr FSCdr) getUniqueID(timezone string) string {
	setupTime, _ := utils.ParseTimeDetectLayout(fsCdr.vars[FS_SETUP_TIME], timezone)
	return utils.Sha1(fsCdr.vars[FS_UUID], setupTime.UTC().String())
}

func (fsCdr FSCdr) getExtraFields() map[string]string {
	extraFields := make(map[string]string, len(fsCdr.cfg.Cdrs.ExtraFields))
	for _, field := range fsCdr.cfg.Cdrs.ExtraFields {
		origFieldVal, foundInVars := fsCdr.vars[field.Id]
		if strings.HasPrefix(field.Id, utils.STATIC_VALUE_PREFIX) { // Support for static values injected in the CDRS. it will show up as {^value:value}
			foundInVars = true
		}
		if !foundInVars {
			origFieldVal = fsCdr.searchExtraField(field.Id, fsCdr.body)
		}
		extraFields[field.Id] = field.ParseValue(origFieldVal)
	}
	return extraFields
}

func (fsCdr FSCdr) searchExtraField(field string, body map[string]interface{}) (result string) {
	for key, value := range body {
		switch v := value.(type) {
		case string:
			if key == field {
				return v
			}
		case float64:
			if key == field {
				return strconv.FormatFloat(v, 'f', -1, 64)
			}
		case map[string]interface{}:
			if result = fsCdr.searchExtraField(field, v); result != "" {
				return
			}
		case []interface{}:
			for _, item := range v {
				if otherMap, ok := item.(map[string]interface{}); ok {
					if result = fsCdr.searchExtraField(field, otherMap); result != "" {
						return
					}
				} else {
					utils.Logger.Warn("Slice with no maps: ", zap.Any("obj", reflect.TypeOf(item)))
				}
			}
		default:
			utils.Logger.Warn("Unexpected type: ", zap.Any("obj", reflect.TypeOf(v)))
		}
	}
	return
}

func (fsCdr FSCdr) AsStoredCdr(timezone string) *CDR {
	storCdr := new(CDR)
	storCdr.UniqueID = fsCdr.getUniqueID(timezone)
	storCdr.ToR = utils.VOICE
	storCdr.OriginID = fsCdr.vars[FS_UUID]
	storCdr.OriginHost = fsCdr.vars[FS_IP]
	storCdr.Source = FS_CDR_SOURCE
	storCdr.RequestType = utils.FirstNonEmpty(fsCdr.vars[utils.CGR_REQTYPE], *fsCdr.cfg.General.DefaultRequestType)
	storCdr.Direction = utils.OUT
	storCdr.Tenant = utils.FirstNonEmpty(fsCdr.vars[utils.CGR_TENANT], *fsCdr.cfg.General.DefaultTenant, utils.USERS)
	storCdr.Category = utils.FirstNonEmpty(fsCdr.vars[utils.CGR_CATEGORY], *fsCdr.cfg.General.DefaultCategory, utils.USERS)
	storCdr.Account = utils.FirstNonEmpty(fsCdr.vars[utils.CGR_ACCOUNT], fsCdr.vars[FS_USERNAME], utils.USERS)
	storCdr.Subject = utils.FirstNonEmpty(fsCdr.vars[utils.CGR_SUBJECT], fsCdr.vars[utils.CGR_ACCOUNT], fsCdr.vars[FS_USERNAME], utils.USERS)
	storCdr.Destination = utils.FirstNonEmpty(fsCdr.vars[utils.CGR_DESTINATION], fsCdr.vars[FS_CALL_DEST_NR], fsCdr.vars[FS_SIP_REQUSER], fsCdr.vars[FS_SIP_TO_USER])
	storCdr.SetupTime, _ = utils.ParseTimeDetectLayout(fsCdr.vars[FS_SETUP_TIME], timezone) // Not interested to process errors, should do them if necessary in a previous step
	pddStr := utils.FirstNonEmpty(fsCdr.vars[FS_PROGRESS_MEDIAMSEC], fsCdr.vars[FS_PROGRESSMS])
	pddStr += "ms"
	storCdr.PDD, _ = time.ParseDuration(pddStr)
	storCdr.AnswerTime, _ = utils.ParseTimeDetectLayout(fsCdr.vars[FS_ANSWER_TIME], timezone)
	storCdr.Usage, _ = utils.ParseDurationWithSecs(fsCdr.vars[FS_DURATION])
	storCdr.Supplier = fsCdr.vars[utils.CGR_SUPPLIER]
	storCdr.DisconnectCause = utils.FirstNonEmpty(fsCdr.vars[utils.CGR_DISCONNECT_CAUSE], fsCdr.vars["hangup_cause"])
	storCdr.ExtraFields = fsCdr.getExtraFields()
	storCdr.Cost = dec.NewVal(-1, 0)
	return storCdr
}
