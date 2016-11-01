package mergemap

import (
	"errors"
	"reflect"
	"strings"
)

var (
	MaxDepth = 32
)

// Merge recursively merges the src and dst maps. Key conflicts are resolved by
// preferring src, or recursively descending, if both src and dst are maps.
func Merge(dst, src map[string]interface{}) (map[string]interface{}, error) {
	return merge(dst, src, 0)
}

func merge(dst, src map[string]interface{}, depth int) (map[string]interface{}, error) {
	if depth > MaxDepth {
		return nil, errors.New("mergemap: maps too deeply nested")
	}

	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			srcMap, srcMapOk := mapify(srcVal)
			dstMap, dstMapOk := mapify(dstVal)
			if srcMapOk && dstMapOk {
				var err error
				srcVal, err = merge(dstMap, srcMap, depth+1)
				if err != nil {
					return nil, err
				}
			} else {
				if !dstMapOk {
					// try convert array with one-field maps to map
					if dstArr, ok := dstVal.([]interface{}); ok {
						for _, item := range dstArr {
							if itemMap, ok := item.(map[string]interface{}); ok {
								if len(itemMap) == 1 {
									for k, v := range itemMap {
										dstMap[k] = v
										dstMapOk = true
									}
								}
							}
						}
					}
				}

				if dstMapOk {
					if srcArr, ok := srcVal.([]interface{}); ok {
						for _, item := range srcArr {
							if itemMap, ok := item.(map[string]interface{}); ok {
								for k, v := range itemMap {
									if dstVal, ok := dstMap[k]; ok {
										vMap, ok := v.(map[string]interface{})
										if !ok {
											vMap = map[string]interface{}{k: v}
										}

										if dstVal, ok := mapify(dstVal); ok {
											vMap, err := merge(dstVal, vMap, depth+1)
											if err != nil {
												return nil, err
											}

											itemMap[k] = vMap
										}
									}
								}
							}
						}
					}
				}
			}
		}

		cleanupMatchingKeys(dst, key)

		dst[key] = srcVal
	}
	return dst, nil
}

func mapify(i interface{}) (map[string]interface{}, bool) {
	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Map {
		m := map[string]interface{}{}
		for _, k := range value.MapKeys() {
			m[k.Interface().(string)] = value.MapIndex(k).Interface()
		}
		return m, true
	}
	return map[string]interface{}{}, false
}

func cleanupMatchingKeys(dest map[string]interface{}, key string) {
	cleanKey := strings.TrimSpace(strings.SplitN(key, "?", 2)[0])

	if cleanKey != key {
		delete(dest, cleanKey)
	}

	for k := range dest {
		if strings.HasPrefix(k, cleanKey+" ") || strings.HasPrefix(k, cleanKey+"?") {
			delete(dest, k)
		}
	}
}
