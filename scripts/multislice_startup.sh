#!/bin/bash

# Change to load a global profile already saved in your Flex
FLEX_GLOBAL_PROFILE=flex_profile_name

# Change to give a specific client name to this application
FLEX_STATION=flex_station_name

./nCAT -headless -profile "${FLEX_GLOBAL_PROFILE}" -station "${FLEX_STATION}" -listen :4532 &
./nCAT -station "${FLEX_STATION}" -slice B -listen :4533 &
./nCAT -station "${FLEX_STATION}" -slice C -listen :4534 &
./nCAT -station "${FLEX_STATION}" -slice D -listen :4535 &