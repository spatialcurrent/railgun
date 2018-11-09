#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

#set -eu

BASEURL='https://tasks.hotosm.org/api/v1/project/5400'
DATASTORE=hotosm_project

go run $DIR/../cmd/railgun/main.go client workspaces add \
--name hotosm \
--title HOTOSM \
--description 'Workspace for Humanitarian OpenStreetMap Team (HOT) data'

go run $DIR/../cmd/railgun/main.go client datastores add \
--workspace hotosm \
--name $DATASTORE \
--title HOT Tasking Manager Project \
--description 'Project on HOT Tasking Manager' \
--uri '"https://tasks.hotosm.org/api/v1/project/" + $project' \
--format json \
--compression ''

go run $DIR/../cmd/railgun/main.go client processes add \
--name project_aoi \
--title Project Area of Interest \
--description 'Get the area of interest for a HOT Tasking Manager project' \
--expression '@areaOfInterest'

go run $DIR/../cmd/railgun/main.go client processes add \
--name project_features \
--title Project Features \
--description 'Get the area of interest for a HOT Tasking Manager project' \
--expression '@features'

go run $DIR/../cmd/railgun/main.go client processes add \
--name project_words \
--title Project Words \
--description 'Get the words used in the description of a HOT Tasking Manager project' \
--expression '(@projectInfo?.description == null) ? [] : (set(split(@projectInfo.description, ` `)) - $irrelevant) | array(@)'

go run $DIR/../cmd/railgun/main.go client processes add \
--name project_hist \
--title Project Histogram \
--description 'Get a histogram of words used in the description of a HOT Tasking Manager project' \
--expression '(@projectInfo?.description == null) ? hist([]) : (hist(split(@projectInfo.description, ` `)) - $irrelevant)'

go run cmd/railgun/main.go client services add \
--name project_words \
--title 'Project Words' \
--description 'Get the words used in the description of a HOT Tasking Manager project' \
--datastore $DATASTORE \
--process project_words

go run cmd/railgun/main.go client services add \
--name project_hist \
--title 'Project Histogram' \
--description 'Get a histogram of words used in the description of a HOT Tasking Manager project' \
--datastore hotosm_project \
--process project_hist \
--defaults '{irrelevant: {"-", OpenStreetMap, OSM, by, to, a, an, and, of, for, in, is, This, project, map}}'

go run cmd/railgun/main.go client services exec \
--service project_words \
--variables '{project: 5400}'

go run cmd/railgun/main.go client services exec \
--service project_hist \
--variables '{project: 5400}'