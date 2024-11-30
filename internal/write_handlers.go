package internal

// currently not implemented. need a better planning. will do it later.
func getNonUniqueNameInfo(sourceId string) string {
	/*
		json data is not accurate. some players names are differently mentioned in different files
		where there are multiple players with same name. need to organise it with proper names and fetch using source_id.
		currently not able to fetch players based on source_id because
		json files contains names inside innings and not source_id.

		eg : search as (2) (3) to get names with issues
	*/
	nonUniqueNames := make(map[string]string)
	nonUniqueNames["d6dcb0c9"] = "Mohammad Shahzad Mohammadi"
	nonUniqueNames["7b9b9aef"] = "Sher Muhammad Abdul Shakoor"
	nonUniqueNames["e191bf68"] = "Ramiz Hasan Raja"

	return nonUniqueNames[sourceId]
}

//func ReplaceNamesInPlayers() {
//
//}
//
//func ReplaceNamesInRegistry() {
//
//}
//
//func ReplaceNamesInInnings() {
//
//}
