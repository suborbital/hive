import { logInfo } from "@suborbital/suborbital"
import { JSON } from "json-as"

// @ts-ignore
@json
class JSONSchema {
    firstName: string
    lastName: string
    age: i32
	meta: Meta
	tags: Array<string>
}

// @ts-ignore
@json
class Meta {
	country: string
}

export function run(_: ArrayBuffer): ArrayBuffer {

	const data: JSONSchema = {
		firstName: 'Connor',
		lastName: 'Hicks',
		age: 26,
		meta: {
			country: "Canada",
		},
		tags: ["hello", "world"]
	}
	
	const stringified = JSON.stringify(data)

	logInfo(stringified)

	// const stringified2 = "{\"firstName\":\"Connor\",\"lastName\":\"Hicks\",\"age\":14,\"meta\":{\"country\":\"Canada\"},\"tags\":[\"hello\",\"world\"]}"
	
	// const parsed = JSON.parse<JSONSchema>(stringified2)

	// const stringifiedAgain = JSON.stringify(parsed)

	return String.UTF8.encode(stringified)
}