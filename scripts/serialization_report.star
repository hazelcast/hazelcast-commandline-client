# This script creates a serialization report
# by traversing all Maps in the cluster and
# getting one of the entries in the map.


def main():
    print("# Map Serialization Report")
    print()

    for map_name in object_list("map"):
        key_set = map_key_set(name=map_name)
        if len(key_set) == 0:
            output(map=map_name)
            continue
        view = map_entry_view(key_set[0], name=map_name)
        output(
            map=map_name,
            key_type=data_type(view["Key"]),
            value_type=data_type(view["Value"]),
        )
