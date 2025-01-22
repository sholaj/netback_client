#!/bin/bash

# Exit on error, unset variable usage, and pipe failure; enable debugging for troubleshooting
set -euo pipefail

# Function to check if the subscription is set
check_subscription() {
    if [ -z "$SUBSCRIPTION" ]; then
        echo "Error: SUBSCRIPTION environment variable is not set."
        exit 1
    fi

    echo "Setting Azure subscription to: $SUBSCRIPTION"
    az account set --subscription "$SUBSCRIPTION" >/dev/null 2>&1
    if [ $? -ne 0 ]; then
        echo "Error: Unable to set subscription to $SUBSCRIPTION. Please check the subscription ID."
        exit 1
    fi
}

# Function to sanitize the cluster name
sanitize_cluster_name() {
    local cluster_name="$1"
    # Remove any surrounding quotes
    cluster_name=$(echo "$cluster_name" | sed 's/^"//;s/"$//')
    echo "$cluster_name"
}

# Function to check if the AKS cluster exists
check_cluster_existence() {
    local cluster_name="$1"
    local resource_group="$2"

    echo "Checking if AKS cluster $cluster_name exists in resource group $resource_group..."
    az aks show --resource-group "$resource_group" --name "$cluster_name" >/dev/null 2>&1
    if [ $? -ne 0 ]; then
        echo "Error: AKS cluster $cluster_name does not exist in resource group $resource_group."
        exit 1
    fi
    echo "Cluster $cluster_name exists."
}

# Function to check the status of the AKS cluster
check_cluster_status() {
    local cluster_name="$1"
    local resource_group="$2"

    echo "Checking status of AKS cluster: $cluster_name in resource group: $resource_group..."

    # Retrieve the cluster's power state
    STATUS=$(az aks show --resource-group "$resource_group" --name "$cluster_name" --query powerState.code -o tsv 2>/dev/null)

    # Check if the az command was successful
    if [ $? -ne 0 ]; then
        echo "Error: Unable to retrieve the status of cluster $cluster_name in resource group $resource_group."
        echo "Please check if the cluster and resource group exist."
        exit 1
    fi

    # Return the cluster status
    echo "$STATUS"
}

# Function to start or stop AKS cluster
manage_aks_cluster() {
    local cluster_name="$1"
    local resource_group="$2"
    local action="$3"

    # Ensure the cluster exists before managing it
    check_cluster_existence "$cluster_name" "$resource_group"

    # Get the current cluster status
    local STATUS
    STATUS=$(check_cluster_status "$cluster_name" "$resource_group")

    if [[ "$action" == "start" ]]; then
        if [[ "$STATUS" == "Running" ]]; then
            echo "Cluster $cluster_name is already started."
            return 0
        fi
        echo "Starting AKS cluster: $cluster_name..."
        az aks start --name "$cluster_name" --resource-group "$resource_group"
    elif [[ "$action" == "stop" ]]; then
        if [[ "$STATUS" == "Stopped" ]]; then
            echo "Cluster $cluster_name is already stopped."
            return 0
        fi
        echo "Stopping AKS cluster: $cluster_name..."
        az aks stop --name "$cluster_name" --resource-group "$resource_group"
    else
        echo "Invalid action: $action. Please specify 'start' or 'stop'."
        exit 1
    fi
}

# Main function
main() {
    local cluster_name="$1"
    local resource_group="$2"
    local action="$3"

    if [[ "$#" -ne 3 ]]; then
        echo "Usage: $0 <cluster_name> <resource_group> <start|stop>"
        exit 1
    fi

    # Check and set the subscription
    check_subscription

    # Sanitize the cluster name
    cluster_name=$(sanitize_cluster_name "$cluster_name")

    # Manage the AKS cluster
    manage_aks_cluster "$cluster_name" "$resource_group" "$action"
}

# Extract command-line arguments
CLUSTER="$1"
RG="$2"
ACTION_FLAG="$3"

# Call the main function
main "$CLUSTER" "$RG" "$ACTION_FLAG"