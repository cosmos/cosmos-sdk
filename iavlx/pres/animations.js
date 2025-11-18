// Reusable animation utilities for IAVL tree visualization

/**
 * Loads an SVG file and injects it into a container element
 * @param {string} svgPath - Path to the SVG file
 * @param {string} containerId - ID of the container element
 * @returns {Promise<SVGElement>} The loaded SVG element
 */
async function loadSVG(svgPath, containerId) {
  const response = await fetch(svgPath);
  const svgText = await response.text();
  const container = document.getElementById(containerId);
  container.innerHTML = svgText;
  return container.querySelector('svg');
}

/**
 * Click-through animation controller for dual SVG (tree + changeset)
 */
class DualSVGAnimator {
  constructor(treeContainerId, changesetContainerId, nodeIds, previousNodes = []) {
    this.treeContainerId = treeContainerId;
    this.changesetContainerId = changesetContainerId;
    this.nodeIds = nodeIds;
    this.previousNodes = previousNodes;
    this.currentIndex = 0;
    this.onUpdate = null;
  }

  /**
   * Initialize the animation state
   */
  init() {
    this.currentIndex = 0;
    this.hideAllChangesetNodes();
    this.showPreviousVersionNodes();
    this.clearTreeAnimations();
    if (this.onUpdate) {
      this.onUpdate(0, this.nodeIds.length);
    }
  }

  /**
   * Show all nodes from previous versions (already written to changeset)
   */
  showPreviousVersionNodes() {
    const changesetContainer = document.getElementById(this.changesetContainerId);
    if (!changesetContainer) return;

    this.previousNodes.forEach(nodeId => {
      const node = changesetContainer.querySelector(`#${nodeId}`);
      if (node) {
        node.style.opacity = '1';
        node.classList.remove('changeset-appear');
      }
    });
  }

  /**
   * Hide all changeset nodes initially
   */
  hideAllChangesetNodes() {
    const changesetContainer = document.getElementById(this.changesetContainerId);
    if (!changesetContainer) return;

    this.nodeIds.forEach(nodeId => {
      const node = changesetContainer.querySelector(`#${nodeId}`);
      if (node) {
        node.style.opacity = '0';
        node.style.transition = 'opacity 0.3s ease-in-out';
      }
    });
  }

  /**
   * Clear all tree animations
   */
  clearTreeAnimations() {
    const treeContainer = document.getElementById(this.treeContainerId);
    if (!treeContainer) return;

    const allNodes = treeContainer.querySelectorAll('.animate, .visited');
    allNodes.forEach(node => {
      node.classList.remove('animate', 'visited');
    });
  }

  /**
   * Advance to next node in sequence
   * @returns {boolean} True if advanced, false if at end
   */
  next() {
    if (this.currentIndex >= this.nodeIds.length) {
      return false;
    }

    const nodeId = this.nodeIds[this.currentIndex];

    // Animate tree node
    this.animateTreeNode(nodeId);

    // Show corresponding changeset node
    this.showChangesetNode(nodeId);

    this.currentIndex++;

    if (this.onUpdate) {
      this.onUpdate(this.currentIndex, this.nodeIds.length);
    }

    return true;
  }

  /**
   * Go back to previous node
   * @returns {boolean} True if went back, false if at beginning
   */
  previous() {
    if (this.currentIndex <= 0) {
      return false;
    }

    this.currentIndex--;
    const nodeId = this.nodeIds[this.currentIndex];

    // Remove animation from tree node
    this.removeTreeNodeAnimation(nodeId);

    // Hide corresponding changeset node
    this.hideChangesetNode(nodeId);

    if (this.onUpdate) {
      this.onUpdate(this.currentIndex, this.nodeIds.length);
    }

    return true;
  }

  /**
   * Animate a tree node with bounce effect
   */
  animateTreeNode(nodeId) {
    const treeContainer = document.getElementById(this.treeContainerId);
    if (!treeContainer) return;

    const node = treeContainer.querySelector(`#${nodeId}`);
    if (node) {
      // Remove previous animation
      node.classList.remove('animate');

      // Trigger reflow
      void node.offsetWidth;

      // Add animation
      node.classList.add('animate');

      // Add visited class after animation
      setTimeout(() => {
        node.classList.add('visited');
      }, 500);
    }
  }

  /**
   * Remove animation from tree node
   */
  removeTreeNodeAnimation(nodeId) {
    const treeContainer = document.getElementById(this.treeContainerId);
    if (!treeContainer) return;

    const node = treeContainer.querySelector(`#${nodeId}`);
    if (node) {
      node.classList.remove('animate', 'visited');
    }
  }

  /**
   * Show a changeset node with elastic grow animation
   */
  showChangesetNode(nodeId) {
    const changesetContainer = document.getElementById(this.changesetContainerId);
    if (!changesetContainer) return;

    const node = changesetContainer.querySelector(`#${nodeId}`);
    if (node) {
      // Remove any existing animation class
      node.classList.remove('changeset-appear');

      // Trigger reflow
      void node.offsetWidth;

      // Add animation class
      node.classList.add('changeset-appear');
      node.style.opacity = '1';
    }
  }

  /**
   * Hide a changeset node
   */
  hideChangesetNode(nodeId) {
    const changesetContainer = document.getElementById(this.changesetContainerId);
    if (!changesetContainer) return;

    const node = changesetContainer.querySelector(`#${nodeId}`);
    if (node) {
      node.classList.remove('changeset-appear');
      node.style.opacity = '0';
    }
  }

  /**
   * Reset to beginning
   */
  reset() {
    this.init();
  }

  /**
   * Check if at the end
   */
  isAtEnd() {
    return this.currentIndex >= this.nodeIds.length;
  }

  /**
   * Check if at the beginning
   */
  isAtBeginning() {
    return this.currentIndex === 0;
  }

  /**
   * Get current progress
   */
  getProgress() {
    return {
      current: this.currentIndex,
      total: this.nodeIds.length,
      percentage: (this.currentIndex / this.nodeIds.length) * 100
    };
  }
}

/**
 * Creates a simple status display
 * @param {string} containerId - ID of the container for status
 * @returns {Object} Status controller with update method
 */
function createStatusDisplay(containerId) {
  const container = document.getElementById(containerId);
  return {
    update: (text) => {
      if (container) {
        container.textContent = text;
      }
    },
    clear: () => {
      if (container) {
        container.textContent = '';
      }
    }
  };
}

// Export for use in presentation
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    loadSVG,
    DualSVGAnimator,
    createStatusDisplay
  };
}