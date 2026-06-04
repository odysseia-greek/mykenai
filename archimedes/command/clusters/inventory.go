package clusters

type cluster struct {
	Name  string
	Nodes []node
}

type node struct {
	Name     string
	Address  string
	User     string
	Identity string
	Arch     string
	Role     string
}

const defaultIdentity = "~/.ssh/id_raspie"

func inventories() []cluster {
	return []cluster{
		{
			Name: "hellas",
			Nodes: []node{
				{Name: "sparta.hellas", Address: "192.168.1.121", User: "pi", Identity: defaultIdentity, Arch: "rpi5", Role: "controller"},
				{Name: "athenai.hellas", Address: "192.168.1.122", User: "pi", Identity: defaultIdentity, Arch: "rpi5", Role: "worker"},
				{Name: "thebai.hellas", Address: "192.168.1.123", User: "pi", Identity: defaultIdentity, Arch: "rpi5", Role: "worker"},
				{Name: "korinthos.hellas", Address: "192.168.1.124", User: "pi", Identity: defaultIdentity, Arch: "rpi5", Role: "worker"},
			},
		},
		{
			Name: "hellenistike",
			Nodes: []node{
				{Name: "pella", Address: "192.168.1.131", User: "pi", Identity: defaultIdentity, Arch: "rpi4", Role: "controller"},
				{Name: "alexandria", Address: "192.168.1.132", User: "pi", Identity: defaultIdentity, Arch: "rpi4", Role: "worker"},
				{Name: "antioch", Address: "192.168.1.133", User: "pi", Identity: defaultIdentity, Arch: "rpi4", Role: "worker"},
			},
		},
	}
}

func selectClusters(name string) []cluster {
	all := inventories()
	if name == "" || name == "all" {
		return all
	}

	for _, candidate := range all {
		if candidate.Name == name {
			return []cluster{candidate}
		}
	}

	return nil
}
