
# NetBackup Ansible Role Design and Implementation

This document provides detailed guidance on designing and implementing an Ansible role for managing NetBackup infrastructure.

## Preparation and Pre-checks

Ensure the required directories, dependencies, and prerequisites are met.

### Tasks
- Create necessary directories for storing configuration files.
- Check the connectivity to the Master Server(s) and validate FQDN reachability.
- Ensure the required ports (e.g., 1556) are open for communication.

```yaml
- name: Create directories for NetBackup configurations
  file:
    path: "/etc/netbackup"
    state: directory
    mode: '0755'

- name: Validate connectivity to Master Servers
  command: ping -c 1 {{ item.api_connection.fqdn }}
  loop: "{{ netbackup_master_servers }}"
  register: ping_results
  failed_when: ping_results.failed
```

---

## Configure Master Server Blocks

Generate configuration files dynamically for each master server using a Jinja2 template.

### Example Template (`master_server_block.j2`)
```jinja
# Master Server Configuration
name: {{ item.name }}
fqdn: {{ item.fqdn }}
api_connection:
  fqdn: {{ item.api_connection.fqdn }}
  port: {{ item.api_connection.port }}
  domain_type: {{ item.api_connection.domain_type }}
  domain_name: {{ item.api_connection.domain_name }}
  user: {{ item.api_connection.user }}
secure_comms:
  ca_fingerprint: {{ item.secure_comms.ca_fingerprint }}
  auth_token: {{ item.secure_comms.auth_token }}
gold_list:
{% for server in item.gold_list %}
  - {{ server }}
{% endfor %}
dmz_media_server_list:
{% for server in item.dmz_client_config.dmz_media_server_list %}
  - {{ server }}
{% endfor %}
```

### Example Task
```yaml
- name: Generate Master Server Configuration Files
  template:
    src: "master_server_block.j2"
    dest: "/etc/netbackup/{{ item.name }}.conf"
    owner: root
    group: root
    mode: '0644'
  loop: "{{ netbackup_master_servers }}"
  notify: Restart NetBackup Service
```

---

## Validate and Apply Secure Communications

Verify secure communication settings, including CA fingerprints and auth tokens.

### Tasks
- Ensure the CA fingerprint matches.
- Verify the `auth_token` by attempting an authenticated API call.

```yaml
- name: Validate CA Fingerprints
  shell: |
    openssl x509 -fingerprint -noout -in /path/to/ca.pem | grep -q "{{ item.secure_comms.ca_fingerprint }}"
  loop: "{{ netbackup_master_servers }}"
  register: ca_validation
  failed_when: ca_validation.stdout.find('Fingerprint mismatch') != -1

- name: Validate Auth Token
  uri:
    url: "https://{{ item.api_connection.fqdn }}:{{ item.api_connection.port }}/api/v1/test"
    method: GET
    headers:
      Authorization: "Bearer {{ item.secure_comms.auth_token }}"
  loop: "{{ netbackup_master_servers }}"
  register: auth_validation
  failed_when: auth_validation.status != 200
```

---

## Handle Dynamic Lists (`gold_list`, `dmz_media_server_list`, etc.)

Ensure `gold_list` and `dmz_media_server_list` are dynamically updated.

### Example Task
```yaml
- name: Ensure gold_list is up-to-date
  lineinfile:
    path: "/etc/netbackup/{{ item.name }}.conf"
    line: "{{ gold_entry }}"
    insertafter: "^gold_list:"
  loop: "{{ netbackup_master_servers }}"
  with_items: "{{ item.gold_list }}"
  when: gold_entry not in item.gold_list
```

---

## Apply Feature Flags

Control the behavior of updates based on feature flags like `it_ensures_repo_config` or `it_manages_server_lists`.

```yaml
- name: Apply gold_list updates if enabled
  lineinfile:
    path: "/etc/netbackup/{{ item.name }}.conf"
    line: "{{ gold_entry }}"
    insertafter: "^gold_list:"
  loop: "{{ netbackup_master_servers }}"
  when: "'it_manages_server_lists' in item and item.it_manages_server_lists"
```

---

## Error Handling

Handle missing or invalid configurations gracefully.

### Tasks
- Check for mandatory fields in `netbackup_master_servers`.
- Log errors for debugging.

```yaml
- name: Check for missing mandatory fields
  debug:
    msg: "Mandatory field missing for {{ item.name }}"
  when: item.name is undefined or item.fqdn is undefined
```

---

## Restart Services

Restart NetBackup services when configuration changes occur.

### Handler
```yaml
- name: Restart NetBackup Service
  service:
    name: netbackup
    state: restarted
```

---

## Validation Scenarios

### Scenario 1: Initial Configuration
- Generate configurations for servers not already configured.
- Validate secure comms (CA fingerprints and auth tokens).

### Scenario 2: Dynamic Updates
- Update dynamic lists like `gold_list` and `dmz_media_server_list` without overwriting existing entries.

### Scenario 3: Feature Flag Behavior
- Enable or disable tasks based on feature flags.

### Scenario 4: Error Scenarios
- Handle and log errors for missing fields or unreachable servers.

---

## Molecule Testing

Use Molecule for automated testing of the role. Define scenarios for:
- Verifying configuration file generation.
- Ensuring server connectivity.
- Testing updates to dynamic lists.

### Molecule Test Playbook
```yaml
---
- name: Test NetBackup Role
  hosts: all
  roles:
    - role: netbackup
```

---

## Example GB Region Configuration

### Sample Code
```yaml
netbackup_config:
  gb_master_servers:
    - name: nbudevtst
      fqdn: nbudevtst.systems.uk.hsbc
      api_connection:
        fqdn: nbudevtst.systems.uk.hsbc
        port: 1556
        domain_type: unixpwd
        domain_name: nbudevtst
        user: nbwebsvc
      secure_comms:
        ca_fingerprint: 95:0B:AB:84...
        auth_token: WJZALDKDKWUXXYRD
      gold_list:
        - gnbuappdev001-cs.systems.uk.hsbc
        - gnbuappdev002-cs.systems.uk.hsbc
      dmz_media_server_list:
        - nbudevtst.systems.uk.hsbc
```

---

To create an Ansible role delivering the feature tested in the provided BDD tests, here’s how you can proceed:

1. Analyze the BDD Test Features

The provided BDD tests appear to test the following functionalities:
	•	Validation and correction of server lists for a master-client relationship.
	•	Ensuring the correct SERVER and MEDIA SERVER entries.
	•	Preserving or updating server lists based on specific configurations (e.g., it_manages_server_lists flag).
	•	Testing compatibility across various OS versions and client versions.
	•	Maintaining custom entries or updating them only when required.

2. Understand the Puppet Feature Requirements

From the analysis:
	•	A “Master Server Block” represents configuration details related to master servers.
	•	Tests verify:
	•	Detection of master server blocks.
	•	Suitability and establishment of master-server relationships based on server and client versions.
	•	Retention or correction of server blocks in existing configurations.

3. Define the Role Structure in Ansible

Ansible roles follow a specific directory structure. Create an Ansible role named manage_server_list:

ansible-galaxy init manage_server_list

This creates the basic structure:

```sh
manage_server_list/
├── tasks/
│   └── main.yml
├── handlers/
├── templates/
├── files/
├── vars/
├── defaults/
├── meta/
└── tests/

```

4. Translate Puppet Logic to Ansible Tasks

Example Task for Validating Server Blocks

Create a task to validate and update server lists. Add this in tasks/main.yml:

---
- name: Validate existing server list
  lineinfile:
    path: "{{ config_file_path }}"
    line: "{{ item }}"
    state: present
  loop: "{{ master_server_blocks }}"
  when: validate_server_list

- name: Add missing entries to server list
  lineinfile:
    path: "{{ config_file_path }}"
    line: "{{ item }}"
    state: present
  loop: "{{ additional_entries }}"
  when: add_missing_entries

- name: Remove invalid entries from server list
  lineinfile:
    path: "{{ config_file_path }}"
    regexp: "{{ invalid_entry_regex }}"
    state: absent
  when: cleanup_invalid_entries

Variables (defaults/main.yml or vars/main.yml)

Define variables based on the tests:

config_file_path: /path/to/server_list.conf
validate_server_list: true
add_missing_entries: true
cleanup_invalid_entries: true

master_server_blocks:
  - "Master Server: master1.example.com"
  - "Media Server: media1.example.com"

additional_entries:
  - "Master Server: master2.example.com"

invalid_entry_regex: '^Invalid.*'

5. Handle Conditionals Based on Flags

Add conditional logic based on the feature flags (it_manages_server_lists, it_ensures_repo_config, etc.):

- name: Check if server list management is enabled
  set_fact:
    manage_server_list: true
  when: it_manages_server_lists == "true"

- name: Ensure repository configuration
  include_tasks: ensure_repo_config.yml
  when: it_ensures_repo_config == "true"

6. Templates for Configuration

Use Jinja2 templates to generate server configuration dynamically:
	•	Create a template in templates/server_list.j2:

{% for server in master_server_blocks %}
Master Server: {{ server }}
{% endfor %}

{% for media in additional_entries %}
Media Server: {{ media }}
{% endfor %}

	•	Add a task to copy this template:

- name: Deploy server list configuration
  template:
    src: server_list.j2
    dest: "{{ config_file_path }}"

7. Test the Role

Use the tests/ directory or run the role in a playbook:

- hosts: all
  roles:
    - role: manage_server_list

8. Iterate Based on Test Cases

---

## Here is a detailed breakdown of tasks and scenarios based on the BDD tests and the provided documentation. The goal is to convert the Puppet logic into an Ansible role while covering all scenarios.

1. Task Breakdown

Each task corresponds to a specific operation described in the BDD tests or the rules defined in the Puppet feature.

### Task 1: Validate Master Server Block

**Purpose:** Verify if a valid “Master Server Block” exists for the client.
- Detect a “Master Server Block” using:
  - FQDN or short name of the master server.
  - Alias of the master server or media server.
  - Secure communication configuration.
- Ensure the block is suitable (client version >= master server version).
- If suitable, mark it as “established.”

**Ansible Implementation:**
```yaml
- name: Validate master server block presence
  set_fact:
    valid_master_blocks: "{{ master_server_blocks | select('search', valid_master_regex) | list }}"
  vars:
    valid_master_regex: ".*Master Server.*"
```

### Task 2: Add Missing Master Server Entries

**Purpose:** Ensure all required entries (Master and Media Servers) are present in the configuration file.
- Check against the list of required servers (master_server_blocks, media_server_blocks).
- Add any missing entries.

**Ansible Implementation:**
```yaml
- name: Add missing master server entries
  lineinfile:
    path: "{{ config_file_path }}"
    line: "{{ item }}"
    state: present
  loop: "{{ missing_master_entries }}"
```

### Task 3: Retain Existing Valid Blocks

**Purpose:** Ensure existing valid “Master Server Blocks” are retained in the original order.
- Retain only valid blocks.
- Keep secure communication-enabled blocks.

**Ansible Implementation:**
```yaml
- name: Retain valid master server blocks
  copy:
    content: |
      {% for block in valid_master_blocks %}
      {{ block }}
      {% endfor %}
    dest: "{{ config_file_path }}"
```

### Task 4: Remove Invalid Blocks

**Purpose:** Remove invalid entries such as:
- Blocks with incorrect syntax.
- Invalid hostnames.
- Non-secure communication blocks.

**Ansible Implementation:**
```yaml
- name: Remove invalid master server blocks
  lineinfile:
    path: "{{ config_file_path }}"
    regexp: "{{ invalid_block_regex }}"
    state: absent
  vars:
    invalid_block_regex: "^Invalid.*"
```

### Task 5: Handle Secure Communication

**Purpose:** Ensure secure communication blocks are configured and retained.
- Add secure communication settings if missing.
- Retain existing secure blocks.

**Ansible Implementation:**
```yaml
- name: Configure secure communication
  lineinfile:
    path: "{{ secure_comm_file }}"
    line: "{{ secure_comm_line }}"
    state: present
```

### Task 6: Ensure Repository Configuration

**Purpose:** If the `it_ensures_repo_config` flag is true, ensure repository configuration is valid.
- Check repository file.
- Update repository settings if needed.

**Ansible Implementation:**
```yaml
- name: Ensure repository configuration
  block:
    - name: Validate repository file
      stat:
        path: "{{ repo_config_file }}"
      register: repo_file_status

    - name: Update repository settings
      lineinfile:
        path: "{{ repo_config_file }}"
        line: "{{ repo_config_line }}"
        state: present
  when: it_ensures_repo_config == "true"
```

### Task 7: Preserve Custom Entries

**Purpose:** Custom entries (not part of the default configuration) should not be overwritten or removed.
- Detect custom entries.
- Retain them during block updates.

**Ansible Implementation:**
```yaml
- name: Retain custom entries
  copy:
    content: |
      {% for entry in custom_entries %}
      {{ entry }}
      {% endfor %}
    dest: "{{ config_file_path }}"
  vars:
    custom_entries: "{{ existing_entries | difference(standard_entries) }}"
```

### Task 8: Update Configuration Based on Flags

**Purpose:** Update the configuration only if specific flags are set:
- `it_manages_server_lists`: Controls whether server lists should be updated.
- `it_ensures_repo_config`: Controls whether the repository should be configured.

**Ansible Implementation:**
```yaml
- name: Check if server list management is enabled
  set_fact:
    manage_server_list: true
  when: it_manages_server_lists == "true"
```

### Task 9: Handle Missing Media Servers

**Purpose:** Add missing media server entries if they do not exist.

**Ansible Implementation:**
```yaml
- name: Add missing media server entries
  lineinfile:
    path: "{{ config_file_path }}"
    line: "{{ item }}"
    state: present
  loop: "{{ media_server_blocks }}"
```

2. Scenarios

Here is a breakdown of the scenarios from the BDD tests and how the tasks address them.

| Scenario | Task | Details |
|----------|------|---------|
| Server list has valid media server entries | Task 2, Task 9 | Ensure all required master and media servers are present. |
| Server list has invalid blocks | Task 4 | Remove invalid entries. |
| Repository configuration is enforced when `it_ensures_repo_config` is true | Task 6 | Check and update repository configuration. |
| Custom entries are preserved | Task 7 | Retain all non-standard (custom) entries during updates. |
| Secure communication blocks are retained | Task 5 | Ensure secure communication is enabled and existing secure blocks are kept. |
| Server list is not updated when `it_manages_server_lists` is false | Task 8 | Skip tasks modifying the server list when this flag is false. |
| Server list updates based on OS and client version compatibility | Task 1, Task 3, Task 2 | Validate master blocks for compatibility, and retain or update as needed. |
| Missing entries are added to the configuration | Task 2, Task 9 | Detect missing entries and append them. |
| Custom configuration is retained when the server has unique or application-specific settings | Task 7 | Preserve custom configurations in the server list. |

3. Steps to Create the Ansible Role

1. **Define Role Structure:**
   Use `ansible-galaxy init manage_server_list` to create a skeleton.

2. **Implement Each Task:**
   - Add the tasks to `tasks/main.yml` or split them into sub-task files for clarity.
   - Use variables for flexibility (e.g., `defaults/main.yml`).

3. **Add Templates:**
   - Create `templates/server_list.j2` for dynamic configuration.
   - Add files for secure communication or repo configurations if needed.

4. **Write Playbook for Testing:**
   Test each scenario using a playbook like:
   ```yaml
   - hosts: all
     roles:
       - manage_server_list
   ```

5. **Iterate Based on Test Outputs:**
   - Compare the outputs with the BDD tests.
   - Debug any failures and update tasks as required.


#######
1. Scenario: Adding a Master Server Block
	•	Goal: Ensure that a new master server block is added correctly if the client doesn’t have a master server.
	•	Actions:
	•	Verify the “Design Master” details from the configuration.
	•	Ensure secure communication setup (e.g., CA fingerprint, auth tokens).
	•	Add the master server block to the client’s configuration.

2. Scenario: Retaining Existing Master Servers
	•	Goal: Existing master servers should remain configured if they are still valid.
	•	Actions:
	•	Check for master servers already listed on the client.
	•	Validate compatibility with the client’s NetBackup version.
	•	Retain master servers in the original order.

3. Scenario: Handling Invalid or Missing Entries
	•	Goal: Detect and fix invalid or incomplete master/media server entries in the client’s configuration.
	•	Actions:
	•	Detect entries with invalid FQDNs or IPs and repair them.
	•	Fill in any missing server details from the central configuration.

4. Scenario: Preserving Custom Server Entries
	•	Goal: Custom configurations on the client must be retained during updates.
	•	Actions:
	•	Identify custom server entries in the client’s configuration.
	•	Ensure these entries are not removed or altered during updates.

5. Scenario: Not Updating Servers Based on Feature Flags
	•	Goal: Skip server list updates if the corresponding feature flag is disabled.
	•	Actions:
	•	Respect the “it_manages_server_lists” feature flag.
	•	Avoid any modifications to the client’s server list if this flag is false.
