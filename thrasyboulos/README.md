# Thrasyboulos

Thrasyboulos is a collection of Kubernetes manifests organized by purpose and environment for the Odysseia-Greek project. The directory structure follows a naming convention inspired by Greek elements.

## Directory Structure

The directory is organized into four main categories, each named after a Greek element:

- **Ge (γῆ, "earth")** - Infrastructure components
  - Contains base infrastructure configurations like databases, message queues, etc.
  - Organized by environment (acc, base, local, prod)

- **Pyr (πῦρ, "fire")** - Components
  - Contains reusable Kubernetes components and building blocks
  - Includes the "koinos" (common) components

- **Aer (ἀήρ, "air")** - Cluster
  - Contains cluster-level configurations and resources
  - Organized by environment (acc, base, local, prod)

- **Hydor (ὕδωρ, "water")** - Applications
  - Contains application manifests and deployments
  - Organized by environment (acc, base, local, prod)

Each category contains configurations for different environments:
- `acc`: Acceptance environment
- `base`: Base configurations (templates)
- `local`: Local development environment
- `prod`: Production environment