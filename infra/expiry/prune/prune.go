package prune

// Make sure there's a -safe flag (that's defaulted to true) where, e.g you are not allowed to terminate VMs that aren't already shut down
// Make sure there's an -action flag (with possible values being 'terminate' & 'shutdown' and default value is 'shutdown') for controlling what action
//  should be taken on expired resources
// Make sure at least one filter flag (region, resource, tag) is provided to avoid catastrophic situations where all resources in an entire
// cloud account are shut-down
// Make sure there's a -yes flag (defaulted to false) that if not set to true will require a confirmation before resources are pruned
