# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0

# e.g
# terraform import observability_object.conn "anzen:cloudConnection|just-a-conn|TENANT|0eb4e853-34fb-4f77-b3fc-b9cd3b462366"
terraform import observability_object.conn "<typeOfObject>|<objectID>|<layerType|<layerID>"