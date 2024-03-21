package model

type Root struct {
	Devices []Device `json:"devices"`
}

type Device struct {
	ID                              string `bson:"_id,omitempty"`
	Name                            string `json:"name"`
	DeviceTypeID                    string `json:"deviceTypeId"`
	Failsafe                        bool   `json:"failsafe"`
	TempMin                         int    `json:"tempMin"`
	TempMax                         int    `json:"tempMax"`
	InstallationPosition            string `json:"installationPosition"`
	InsertInto19InchCabinet         bool   `json:"insertInto19InchCabinet"`
	MotionEnable                    bool   `json:"motionEnable"`
	SiplusCatalog                   bool   `json:"siplusCatalog"`
	SimaticCatalog                  bool   `json:"simaticCatalog"`
	RotationAxisNumber              int    `json:"rotationAxisNumber"`
	PositionAxisNumber              int    `json:"positionAxisNumber"`
	AdvancedEnvironmentalConditions bool   `json:"advancedEnvironmentalConditions,omitempty"`
	TerminalElement                 bool   `json:"terminalElement,omitempty"`
}
