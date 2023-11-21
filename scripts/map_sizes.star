# This script prints sizes of all maps in the cluster

def main():
    print("# Map Size Report")
    print()

    for map_name in object_list("map"):
        size = map_size(name=map_name)
        output(name=map_name, map_size=size)
