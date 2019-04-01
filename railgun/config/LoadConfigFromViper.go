package config

import (
	"reflect"
)

import (
	"github.com/spatialcurrent/viper"
)

func LoadConfigFromViper(c interface{}, v *viper.Viper) {

	configValue := reflect.ValueOf(c).Elem()
	configType := reflect.TypeOf(c).Elem()

	for i := 0; i < configValue.NumField(); i++ {
		structField := configType.Field(i)
		fieldValue := configValue.FieldByName(structField.Name)
		fieldType := structField.Type
		if fieldType.Kind() == reflect.Ptr {
			if fieldType.Elem().Kind() == reflect.Struct {
				LoadConfigFromViper(fieldValue.Interface(), v)
			}
		} else {
			if key, ok := structField.Tag.Lookup("viper"); ok && key != "" && key != "-" {
				switch fieldType.Kind() {
				case reflect.String:
					fieldValue.SetString(v.GetString(key))
				case reflect.Bool:
					fieldValue.SetBool(v.GetBool(key))
				case reflect.Int:
					fieldValue.SetInt(int64(v.GetInt(key)))
				case reflect.Slice:
					if fieldType.Elem().Kind() == reflect.String {
						fieldValue.Set(reflect.ValueOf(v.GetStringSlice(key)))
					}
				default:
					if fieldType.Name() == "Duration" {
						fieldValue.Set(reflect.ValueOf(v.GetDuration(key)))
					}
				}
			}
		}
	}
}
