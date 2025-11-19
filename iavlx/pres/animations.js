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

/**
 * Compaction animation controller
 */
class CompactionAnimator {
  constructor(originalContainerId, compactedContainerId, allNodes, copiedNodes) {
    this.originalContainerId = originalContainerId;
    this.compactedContainerId = compactedContainerId;
    this.allNodes = allNodes;
    this.copiedNodes = new Set(copiedNodes);
    this.currentIndex = 0;
    this.onUpdate = null;
  }

  /**
   * Initialize the animation state
   */
  init() {
    this.currentIndex = 0;
    this.hideAllCompactedNodes();
    if (this.onUpdate) {
      this.onUpdate(0, this.allNodes.length);
    }
  }

  /**
   * Hide all compacted changeset nodes initially
   */
  hideAllCompactedNodes() {
    const compactedContainer = document.getElementById(this.compactedContainerId);
    if (!compactedContainer) return;

    this.copiedNodes.forEach(nodeId => {
      const node = compactedContainer.querySelector(`#${nodeId}`);
      if (node) {
        node.style.opacity = '0';
      }
    });
  }

  /**
   * Advance to next node in sequence
   */
  next() {
    if (this.currentIndex >= this.allNodes.length) {
      return false;
    }

    const nodeId = this.allNodes[this.currentIndex];
    const isCopied = this.copiedNodes.has(nodeId);

    if (isCopied) {
      // Node is copied - show it in compacted changeset
      this.showCompactedNode(nodeId);
      this.highlightOriginalNode(nodeId);
    } else {
      // Node is pruned - mark it as pruned in original
      this.markPrunedNode(nodeId);
    }

    this.currentIndex++;

    if (this.onUpdate) {
      this.onUpdate(this.currentIndex, this.allNodes.length);
    }

    return true;
  }

  /**
   * Go back to previous node
   */
  previous() {
    if (this.currentIndex <= 0) {
      return false;
    }

    this.currentIndex--;
    const nodeId = this.allNodes[this.currentIndex];
    const isCopied = this.copiedNodes.has(nodeId);

    if (isCopied) {
      this.hideCompactedNode(nodeId);
      this.unhighlightOriginalNode(nodeId);
    } else {
      this.unmarkPrunedNode(nodeId);
    }

    if (this.onUpdate) {
      this.onUpdate(this.currentIndex, this.allNodes.length);
    }

    return true;
  }

  /**
   * Show a node in the compacted changeset
   */
  showCompactedNode(nodeId) {
    const compactedContainer = document.getElementById(this.compactedContainerId);
    if (!compactedContainer) return;

    const node = compactedContainer.querySelector(`#${nodeId}`);
    if (node) {
      node.classList.remove('changeset-appear');
      void node.offsetWidth;
      node.classList.add('changeset-appear');
      node.style.opacity = '1';
    }
  }

  /**
   * Hide a node in the compacted changeset
   */
  hideCompactedNode(nodeId) {
    const compactedContainer = document.getElementById(this.compactedContainerId);
    if (!compactedContainer) return;

    const node = compactedContainer.querySelector(`#${nodeId}`);
    if (node) {
      node.classList.remove('changeset-appear');
      node.style.opacity = '0';
    }
  }

  /**
   * Highlight a copied node in original
   */
  highlightOriginalNode(nodeId) {
    const originalContainer = document.getElementById(this.originalContainerId);
    if (!originalContainer) return;

    const node = originalContainer.querySelector(`#${nodeId}`);
    if (node) {
      node.classList.add('animate');
      setTimeout(() => {
        node.classList.remove('animate');
        node.classList.add('visited');
      }, 500);
    }
  }

  /**
   * Unhighlight a node in original
   */
  unhighlightOriginalNode(nodeId) {
    const originalContainer = document.getElementById(this.originalContainerId);
    if (!originalContainer) return;

    const node = originalContainer.querySelector(`#${nodeId}`);
    if (node) {
      node.classList.remove('animate', 'visited');
    }
  }

  /**
   * Mark a node as pruned (turn black/gray)
   */
  markPrunedNode(nodeId) {
    const originalContainer = document.getElementById(this.originalContainerId);
    if (!originalContainer) return;

    const node = originalContainer.querySelector(`#${nodeId}`);
    if (node) {
      // Find all child elements (ellipse, polygon, text) and make them gray
      const shapes = node.querySelectorAll('ellipse, polygon, rect');
      const texts = node.querySelectorAll('text');

      shapes.forEach(shape => {
        shape.setAttribute('data-original-fill', shape.getAttribute('fill'));
        shape.setAttribute('fill', '#666');
        shape.setAttribute('stroke', '#333');
      });

      texts.forEach(text => {
        text.setAttribute('data-original-fill', text.getAttribute('fill'));
        text.setAttribute('fill', '#999');
      });
    }
  }

  /**
   * Unmark a pruned node (restore original color)
   */
  unmarkPrunedNode(nodeId) {
    const originalContainer = document.getElementById(this.originalContainerId);
    if (!originalContainer) return;

    const node = originalContainer.querySelector(`#${nodeId}`);
    if (node) {
      const shapes = node.querySelectorAll('ellipse, polygon, rect');
      const texts = node.querySelectorAll('text');

      shapes.forEach(shape => {
        const originalFill = shape.getAttribute('data-original-fill');
        if (originalFill) {
          shape.setAttribute('fill', originalFill);
          shape.removeAttribute('data-original-fill');
        }
        shape.removeAttribute('stroke');
      });

      texts.forEach(text => {
        const originalFill = text.getAttribute('data-original-fill');
        if (originalFill) {
          text.setAttribute('fill', originalFill);
          text.removeAttribute('data-original-fill');
        }
      });
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
    return this.currentIndex >= this.allNodes.length;
  }

  /**
   * Check if at the beginning
   */
  isAtBeginning() {
    return this.currentIndex === 0;
  }
}

/**
 * Navigation animation controller for tree traversal
 * Shows path from root to target leaf, highlighting nodes in both tree and changeset
 */
class NavigationAnimator {
  constructor(treeContainerId, changesetContainerId, navigationPath) {
    this.treeContainerId = treeContainerId;
    this.changesetContainerId = changesetContainerId;
    this.navigationPath = navigationPath;
    this.currentIndex = 0;
    this.onUpdate = null;
  }

  /**
   * Initialize the animation state
   */
  init() {
    this.currentIndex = 0;
    this.resetAllNodes();
    if (this.onUpdate) {
      this.onUpdate(0, this.navigationPath.length);
    }
  }

  /**
   * Reset all nodes to normal size
   */
  resetAllNodes() {
    const treeContainer = document.getElementById(this.treeContainerId);
    const changesetContainer = document.getElementById(this.changesetContainerId);

    if (treeContainer) {
      const allNodes = treeContainer.querySelectorAll('g[id]');
      allNodes.forEach(node => {
        node.classList.remove('nav-highlight');
      });
    }

    if (changesetContainer) {
      const allNodes = changesetContainer.querySelectorAll('g[id]');
      allNodes.forEach(node => {
        node.classList.remove('nav-highlight');
      });
    }
  }

  /**
   * Advance to next node in navigation path
   */
  next() {
    if (this.currentIndex >= this.navigationPath.length) {
      return false;
    }

    // Unhighlight previous node if there was one
    if (this.currentIndex > 0) {
      const prevNodeId = this.navigationPath[this.currentIndex - 1];
      this.unhighlightNode(this.treeContainerId, prevNodeId);
      this.unhighlightNode(this.changesetContainerId, prevNodeId);
    }

    const nodeId = this.navigationPath[this.currentIndex];

    // Highlight current node in both tree and changeset
    this.highlightNode(this.treeContainerId, nodeId);
    this.highlightNode(this.changesetContainerId, nodeId);

    this.currentIndex++;

    if (this.onUpdate) {
      this.onUpdate(this.currentIndex, this.navigationPath.length);
    }

    return true;
  }

  /**
   * Go back to previous node
   */
  previous() {
    if (this.currentIndex <= 0) {
      return false;
    }

    // Unhighlight current node
    const currentNodeId = this.navigationPath[this.currentIndex - 1];
    this.unhighlightNode(this.treeContainerId, currentNodeId);
    this.unhighlightNode(this.changesetContainerId, currentNodeId);

    this.currentIndex--;

    // Highlight previous node if we're not at the beginning
    if (this.currentIndex > 0) {
      const prevNodeId = this.navigationPath[this.currentIndex - 1];
      this.highlightNode(this.treeContainerId, prevNodeId);
      this.highlightNode(this.changesetContainerId, prevNodeId);
    }

    if (this.onUpdate) {
      this.onUpdate(this.currentIndex, this.navigationPath.length);
    }

    return true;
  }

  /**
   * Highlight a node (make it bigger and keep it big)
   */
  highlightNode(containerId, nodeId) {
    const container = document.getElementById(containerId);
    if (!container) return;

    const node = container.querySelector(`#${nodeId}`);
    if (node) {
      node.classList.add('nav-highlight');
    }
  }

  /**
   * Remove highlight from a node
   */
  unhighlightNode(containerId, nodeId) {
    const container = document.getElementById(containerId);
    if (!container) return;

    const node = container.querySelector(`#${nodeId}`);
    if (node) {
      node.classList.remove('nav-highlight');
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
    return this.currentIndex >= this.navigationPath.length;
  }

  /**
   * Check if at the beginning
   */
  isAtBeginning() {
    return this.currentIndex === 0;
  }
}

// Export for use in presentation
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    loadSVG,
    DualSVGAnimator,
    CompactionAnimator,
    NavigationAnimator,
    createStatusDisplay
  };
}