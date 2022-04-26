# Thing Model

## Attributes

### Volume

| property   | value             |
|------------|-------------------|
| type       | volume            |
| permission | read/write/notify |
| value_type | int               |

### On Off

| property    | value             |
|-------------|-------------------|
| type        | on_off            |
| permission  | read/write/notify |
| val_type    | string            |
| valid_value | on/off/switch     |

### Brightness

| property    | value             |
|-------------|-------------------|
| type        | brightness        |
| permission  | read/write/notify |
| val_type    | int               |
| valid_value | 1-100             |

### Color Temperature

| property   | value             |
|------------|-------------------|
| type       | color_temp        |
| permission | read/write/notify |
| val_type   | int               |

### RGB

| property    | value             |
|-------------|-------------------|
| type        | rgb               |
| permission  | read/write/notify |
| val_type    | string            |
| valid_value | RGB Hex           |

### Model

| property   | value  |
|------------|--------|
| type       | model  |
| permission | read   |
| val_type   | string |

### Manufacturer

| property   | value        |
|------------|--------------|
| type       | manufacturer |
| permission | read         |
| val_type   | string       |

### Identify

| property   | value    |
|------------|----------|
| type       | identify |
| permission | read     |
| val_type   | string   |

### Version

| property   | value   |
|------------|---------|
| type       | version |
| permission | read    |
| val_type   | string  |

### Name

| property   | value  |
|------------|--------|
| type       | name   |
| permission | read   |
| value_type | string |

### Current Position

| property   | value            |
|------------|------------------|
| type       | current_position |
| permission | read/notify      |
| value_type | int              |
| value      | 0-100            |

### Target Position

| property   | value             |
|------------|-------------------|
| type       | target_position   |
| permission | read/write/notify |
| value_type | int               |
| value      | 0-100             |

### State

| property    | value       |
|-------------|-------------|
| type        | state       |
| permission  | read/notify |
| value_type  | int         |
| valid_value | 0/1/2       |

### Style

| property    | value                |
|-------------|----------------------|
| type        | contact_sensor_state |
| permission  | read/notify          |
| value_type  | int                  |
| valid_value | 0/1/2/3              |

### Direction

| property    | value       |
|-------------|-------------|
| type        | direction   |
| permission  | read/notify |
| value_type  | int         |
| valid_value | 0/1         |

### Upper Limit

| property    | value             |
|-------------|-------------------|
| type        | upper_limit       |
| permission  | read/write/notify |
| value_type  | int               |
| valid_value | 0/1               |

### Lower Limit

| property    | value             |
|-------------|-------------------|
| type        | lower_limit       |
| permission  | read/write/notify |
| value_type  | int               |
| valid_value | 0/1               |

### Humidity

| property   | value       |
|------------|-------------|
| type       | humidity    |
| permission | read/notify |
| value_type | int         |

### Temperature

| property   | value       |
|------------|-------------|
| type       | temperature |
| permission | read/notify |
| value_type | float       |

### Contact Sensor State

| property   | value                |
|------------|----------------------|
| type       | contact_sensor_state |
| permission | read/write/notify    |
| value_type | int                  |

### Leak Detected

| property   | value         |
|------------|---------------|
| type       | leak_detected |
| permission | read/notify   |
| value_type | int           |

### Switch Event

| property   | value             |
|------------|-------------------|
| type       | switch_event      |
| permission | read/write/notify |
| value_type | int               |

### Target State

| property   | value             |
|------------|-------------------|
| type       | target_state      |
| permission | read/write/notify |
| value_type | int               |

### Current State

| property   | value         |
|------------|---------------|
| type       | current_state |
| permission | read/notify   |
| value_type | int           |

### Motion Detected

| property   | value           |
|------------|-----------------|
| type       | motion_detected |
| permission | read/notify     |
| value_type | bool            |

### Battery

| property   | value       |
|------------|-------------|
| type       | battery     |
| permission | read/notify |
| value_type | float       |

### Lock Target State

| property   | value             |
|------------|-------------------|
| type       | lock_target_state |
| permission | read/write/notify |
| value_type | int               |

### Logs

| property   | value             |
|------------|-------------------|
| type       | logs              |
| permission | read/write/notify |
| value_type | string            |

### Active

| property   | value       |
|------------|-------------|
| type       | active      |
| permission | read/notify |
| value_type | int         |

### Current Temperature

| property   | value               |
|------------|---------------------|
| type       | current_temperature |
| permission | read/notify         |
| value_type | float               |

### Current Heating Cooling State

| property   | value                         |
|------------|-------------------------------|
| type       | current_heating_cooling_state |
| permission | read/notify                   |
| value_type | int                           |

### Target Heating Cooling State

| property   | value                        |
|------------|------------------------------|
| type       | target_heating_cooling_state |
| permission | read/write/notify            |
| value_type | int                          |

### Heating Threshold Temperature

| property   | value                         |
|------------|-------------------------------|
| type       | heating_threshold_temperature |
| permission | read/write/notify             |
| value_type | int                           |

### Cooling Threshold Temperature

| property   | value                         |
|------------|-------------------------------|
| type       | cooling_threshold_temperature |
| permission | read/write/notify             |
| value_type | int                           |

### Current Heater Cooler State

| property   | value                       |
|------------|-----------------------------|
| type       | current_heater_cooler_state |
| permission | read/notify                 |
| value_type | int                         |

### Target Heater Cooler State

| property   | value                      |
|------------|----------------------------|
| type       | target_heater_cooler_state |
| permission | read/write/notify          |
| value_type | int                        |

### Rotation Speed

| property   | value             |
|------------|-------------------|
| type       | rotation_speed    |
| permission | read/write/notify |
| value_type | int               |

### Swing Mode

| property   | value             |
|------------|-------------------|
| type       | swing_mode        |
| permission | read/write/notify |
| value_type | int               |

### Permit Join

| property   | value             |
|------------|-------------------|
| type       | permit_join       |
| permission | read/write/notify |
| value_type | int               |

### Alert

| property   | value |
|------------|-------|
| type       | alert |
| permission ||
| value_type | int   |

### Status Low Battery

| property   | value              |
|------------|--------------------|
| type       | status_low_battery |
| permission | read/write/notify  |
| value_type | int                |

### Contact Sensor State

| property   | value                |
|------------|----------------------|
| type       | contact_sensor_state |
| permission | read/notify          |
| value_type | int                  |

## Services

### Switch

| property   | value             |
|------------|-------------------|
| type       | switch            |
| attributes | [On Off](#on-off) |

### Outlet

| property   | value             |
|------------|-------------------|
| type       | outlet            |
| attributes | [On Off](#on-off) |

### Light Bulb

| property   | value                                                                                                          |
|------------|----------------------------------------------------------------------------------------------------------------|
| type       | light_bulb                                                                                                     |
| attributes | [On Off](#on-off) <br> [Brightness](#brightness) <br> [Color Temperature](#color-temperature) <br> [RGB](#rgb) |

### Gateway

| property   | value             |
|------------|-------------------|
| type       | gateway           |
| attributes | [On Off](#on-off) |

### Curtain

| property   | value                                                                                                                                                                                                                  |
|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| type       | gateway                                                                                                                                                                                                                |
| attributes | [Current Position](#current-position) <br> [Target Position](#target-position) <br> [State](#state) <br> [Style](#style) <br>[Direction](#direction) <br> [Upper Limit](#upper-limit) <br> [Lower Limit](#lower-limit) |

### Contact Sensor

| property   | value                                                                  |
|------------|------------------------------------------------------------------------|
| type       | contact_sensor                                                         |
| attributes | [Contact Sensor State](#contact-sensor-state) <br> [Battery](#battery) |

### Temp Humidity Sensor

| property   | value                                                                           |
|------------|---------------------------------------------------------------------------------|
| type       | temperature_humidity_sensor                                                     |
| attributes | [Temperature](#temperature) <br> [Humidity](#humidity) <br> [Battery](#battery) |

### Stateless Switch

| property   | value                                                 |
|------------|-------------------------------------------------------|
| type       | stateless_switch                                      |
| attributes | [SwitchEvent](#switch-event) <br> [Battery](#battery) |

### Security System

| property   | value                                                            |
|------------|------------------------------------------------------------------|
| type       | security_system                                                  |
| attributes | [TargetState](#target-state) <br> [CurrentState](#current-state) |

### MotionSensor

| property   | value                                                        |
|------------|--------------------------------------------------------------|
| type       | motion_sensor                                                |
| attributes | [Motion Detected](#motion-detected) <br> [Battery](#battery) |

### Lock

| property   | value                                                                             |
|------------|-----------------------------------------------------------------------------------|
| type       | lock                                                                              |
| attributes | [LockTargetState](#lock-target-state) <br> [Logs](#logs) <br> [Battery](#battery) |

### HeaterCooler

| property   | value                                                                                                                                                                                                                                                                                                                                                                                                              |
|------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| type       | heater_cooler                                                                                                                                                                                                                                                                                                                                                                                                      |
| attributes | [Active](#active) <br> [Current Temperature](#current-temperature) <br> [Cooling Threshold Temperature](#cooling-threshold-temperature) <br> [Heating Threshold Temperature](#heating-threshold-temperature) <br> [Current Heater Cooler State](#current-heater-cooler-state) <br> [Target Heater Cooler State](#target-heater-cooler-state) <br> [Rotation Speed](#rotation-speed) <br> [Swing Mode](#swing-mode) |






