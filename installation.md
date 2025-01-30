---

# **NetBackup Installation and Configuration on Linux**

This guide provides step-by-step instructions for installing NetBackup on Linux, configuring the client, and testing the setup to ensure proper functionality.

---

## **1. Installing NetBackup on Linux**

### **Prerequisites**
- Ensure your Linux system meets the hardware and software requirements for NetBackup.
- Verify that the system has the necessary dependencies installed (e.g., `libstdc++`, `glibc`, etc.).
- Download the NetBackup installation package from the Veritas website or your organization's software repository.

### **Installation Steps**
1. Extract the installation package:
   ```bash
   tar -xvf netbackup_client_linux.tar.gz
   ```

2. Navigate to the extracted directory:
   ```bash
   cd netbackup_client_linux
   ```

3. Run the installation script:
   ```bash
   ./install
   ```

4. Follow the on-screen prompts to complete the installation.

---

## **2. Configuring the NetBackup Client**

The NetBackup client configuration file is located at `/usr/openv/netbackup/bp.conf`. Below is an example configuration:

```ini
# NetBackup Client Configuration File (bp.conf)

SERVER = netbackup_server_hostname
CLIENT_NAME = client_hostname
EMMSERVER = netbackup_server_hostname
# Optional: Specify the media server if different from the master server
MEDIA_SERVER = media_server_hostname
# Optional: Specify the storage unit
STORAGE_UNIT = Default
# Optional: Specify the backup policy
POLICY = DailyBackup
# Optional: Specify the backup schedule
SCHEDULE = FullBackup
# Optional: Enable or disable compression
COMPRESSION = ON
# Optional: Specify the backup type (e.g., FULL, INCR)
BACKUP_TYPE = FULL
# Optional: Specify the number of retries
RETRIES = 3
# Optional: Specify the retry interval in seconds
RETRY_INTERVAL = 300
# Optional: Specify the log level (0-5)
LOG_LEVEL = 3
# Optional: Specify the log file location
LOG_FILE = /var/log/netbackup.log
```

### **Key Configuration Parameters**
- **SERVER**: The hostname or IP address of the NetBackup master server.
- **CLIENT_NAME**: The hostname of the client machine.
- **EMMSERVER**: The hostname or IP address of the Enterprise Media Manager (EMM) server.
- **MEDIA_SERVER**: The hostname or IP address of the media server (if different from the master server).
- **STORAGE_UNIT**: The storage unit to use for backups.
- **POLICY**: The backup policy to apply.
- **SCHEDULE**: The backup schedule to follow.
- **COMPRESSION**: Enable or disable compression for backups.
- **BACKUP_TYPE**: The type of backup (e.g., FULL, INCR).
- **RETRIES**: The number of retry attempts for failed backups.
- **RETRY_INTERVAL**: The time interval (in seconds) between retries.
- **LOG_LEVEL**: The verbosity level for logging (0 = minimal, 5 = verbose).
- **LOG_FILE**: The location of the log file.

---

## **3. Testing the NetBackup Configuration**

### **Verify NetBackup Services**
1. Check the status of NetBackup services:
   ```bash
   /etc/init.d/netbackup status
   ```

2. Start the services if they are not running:
   ```bash
   /etc/init.d/netbackup start
   ```

### **Test Communication with the NetBackup Server**
1. Test connectivity to the NetBackup server:
   ```bash
   /usr/openv/netbackup/bin/bpclntcmd -pn
   ```

2. Test hostname resolution:
   ```bash
   /usr/openv/netbackup/bin/bpclntcmd -hn
   ```

### **Validate the Configuration File**
1. Check the configuration file:
   ```bash
   cat /usr/openv/netbackup/bp.conf
   ```

2. Test the configuration file:
   ```bash
   /usr/openv/netbackup/bin/bpclntcmd -ip <netbackup_server_hostname>
   ```

### **Perform a Manual Backup**
1. Initiate a manual backup:
   ```bash
   /usr/openv/netbackup/bin/bpbackup -p <policy_name>
   ```

2. Monitor the backup progress:
   ```bash
   /usr/openv/netbackup/bin/bpdbm -cmd -get_state
   ```

3. Check the backup logs:
   ```bash
   tail -f /usr/openv/netbackup/logs/bpbackup.log
   ```

### **Verify Backup Completion**
1. Check the backup status:
   ```bash
   /usr/openv/netbackup/bin/bpdbm -cmd -get_state
   ```

2. View detailed job logs:
   ```bash
   /usr/openv/netbackup/bin/bpdbjobs
   ```

3. Check the NetBackup Activity Monitor in the web interface.

### **Test Restore Functionality**
1. Restore a file or directory:
   ```bash
   /usr/openv/netbackup/bin/bprestore -p <policy_name> -C <client_name> -s <server_name> -t <date> -D /path/to/restore -R /path/to/restore
   ```

2. Verify that the restored files match the original data.

### **Check for Errors**
1. NetBackup client logs:
   ```bash
   /usr/openv/netbackup/logs/
   ```

2. System logs:
   ```bash
   /var/log/messages
   ```

---

## **4. Troubleshooting Common Issues**
- **Connection Issues**:
  - Verify network connectivity between the client and server.
  - Ensure that firewalls are not blocking NetBackup ports (e.g., 1556, 13724).
  - Check DNS resolution for both client and server.

- **Policy Issues**:
  - Ensure that the backup policy is correctly configured on the NetBackup server.
  - Verify that the client is included in the policy.

- **Permission Issues**:
  - Ensure that the NetBackup client has the necessary permissions to access the files and directories being backed up.

---

