package main

func get12Articles() []ContentSeed {
	var all []ContentSeed
	all = append(all, getArticles1to6()...)
	all = append(all, getArticles7to12()...)
	return all
}
