import json
from nanoid import generate

path = "class-input.json"

with open(path) as json_file:
    data = json.load(json_file)


keynames = {
    "textbook": {"NOP": "page", "INDEXABLE": "text"},
    "syllabus": {"NOP": "line", "INDEXABLE": "text"},
    "faq": {"NOP": "answer", "INDEXABLE": "question"},
}

to_ingest = {}
for key, list_items in data.items():
    for item in list_items:
        exp_key, indexab_key = keynames[key].values()
        indexab = item[indexab_key]

        if key not in to_ingest:
            to_ingest[key] = []

        new_id = "id:" + generate(size=10)
        to_ingest[key] += [{
            new_id: indexab, 
            "key": new_id, 
            "location": exp_key,
            "indexed_text": indexab,
            "value": str(item[exp_key])
        }]

with open('to_ingest.json', 'w') as outfile:
    json.dump(to_ingest, outfile, indent=4)
