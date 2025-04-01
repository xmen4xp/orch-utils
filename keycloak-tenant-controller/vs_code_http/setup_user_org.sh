# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

user="all-groups-example-user"
client_name="system-client"
domain="kind.internal:443"
#domain="cluster.onprem:443"
token=$(
curl -k -s \
  --request POST \
  --header "Content-Type: application/x-www-form-urlencoded" \
  --data "username=$user&password=xxxx&grant_type=password&client_id=$client_name&scope=openid" \
  https://keycloak.$domain/realms/master/protocol/openid-connect/token  | jq -r '.access_token'
)


## Create org
curl -k -s \
  --request PUT \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $token" \
  --data "{\"description\": \"Test Organization\"}"\
  https://api.$domain/v1/orgs/ci-org

echo "sleep 10"
sleep 10

## Get org id for ci-org
org_id="$(kubectl get RuntimeOrg -o yaml | grep -e "nexus/display_name:" -e "uid" | grep -A 1 ci-org | sed "s/ * / /g" | cut -d " " -f 3 |tail -n 1)"
echo "org_id: $org_id"
group=$org_id"_project-manager-group"
echo $group

## get id for project-manager-group

group_id=$(
curl -k -s \
  --request GET \
  --header "Authorization: Bearer $token" \
  https://keycloak.$domain/admin/realms/master/groups/ | jq ".[] | select( .name == \"$group\" ).id"
)
group_id=$(echo $group_id |sed "s/\"//g")
echo "Group id: $group_id"

### get id for $user
user_id=$(
curl -k -s \
  --request GET \
  --header "Authorization: Bearer $token" \
  https://keycloak.$domain/admin/realms/master/users/ | jq ".[] | select( .username == \"$user\" ).id"
)


user_id=$(echo $user_id |sed "s/\"//g")
echo "User id $user_id"


# add $user to group 
curl -k -s \
  --request PUT \
  --header "Accept: application/json" \
  --header "Authorization: Bearer $token" \
  https://keycloak.$domain/admin/realms/master/users/$user_id/groups/$group_id


## Now we can create the project

# we need to update JWT to have the group access added above
token=$(
curl -k -s \
  --request POST \
  --header "Content-Type: application/x-www-form-urlencoded" \
  --data "username=$user&password=xxxx&grant_type=password&client_id=$client_name&scope=openid" \
  https://keycloak.$domain/realms/master/protocol/openid-connect/token  | jq -r '.access_token'
)

echo "sleep 10"
sleep 10
## create project
curl -k -s \
  --request PUT \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $token" \
  --data "{\"description\": \"Test Project\"}"\
  https://api.$domain/v1/projects/ci-project

## get project uid
project_id="$(kubectl get RuntimeProject -o yaml | grep -e "nexus/display_name:" -e "uid" | grep -A 1 ci-project | sed "s/ * / /g" | cut -d " " -f 3 |tail -n 1)"

echo "project_id: $project_id"
project_member_role=$org_id"_"$project_id"_member-role"
echo "project Role: "$project_member_role


project_member_role_keycloak_id=$(
curl -k -s \
  --request GET \
  --header "Authorization: Bearer $token" \
  https://keycloak.$domain/admin/realms/master/roles | jq ".[] | select( .name == \"$project_member_role\" ).id"
)

echo "project_member_role_keycloak_id: $project_member_role_keycloak_id"
project_member_role_keycloak_id=$(echo $project_member_role_keycloak_id |sed "s/\"//g")


client_id=$(
curl -k -s \
  --request GET \
  --header "Authorization: Bearer $token" \
  https://keycloak.$domain/admin/realms/master/clients | jq ".[] | select( .clientId == \"$client_name\" ).id"
)

client_id=$(echo $client_id |sed "s/\"//g")
echo $client_id



## add Role <orgid>_<project_id>_member-role to user
curl -k -s \
  --request POST \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $token"\
  --data "[{\"id\":\"$project_member_role_keycloak_id\",\"name\":\"$project_member_role\"}]"\
   https://keycloak.$domain/admin/realms/master/users/$user_id/role-mappings/realm


# Finally refresh the JWT token so that token has access to <orgid>_<project_id>_member-role
token=$(
curl -k -s \
  --request POST \
  --header "Content-Type: application/x-www-form-urlencoded" \
  --data "username=$user&password=xxxx&grant_type=password&client_id=$client_name&scope=openid" \
  https://keycloak.$domain/realms/master/protocol/openid-connect/token  | jq -r '.access_token'
)

## 