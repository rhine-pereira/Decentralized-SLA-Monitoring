// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract SLARootRegistry {

    struct RootData {
        bytes32 root;
        uint256 timestamp;
    }

    mapping(uint256 => RootData) public roots;

    event RootStored(uint256 indexed bucketId, bytes32 root);

    function storeRoot(uint256 bucketId, bytes32 root) external {
        require(roots[bucketId].root == bytes32(0), "Already exists");

        roots[bucketId] = RootData({
            root: root,
            timestamp: block.timestamp
        });

        emit RootStored(bucketId, root);
    }

    function getRoot(uint256 bucketId) external view returns (bytes32) {
        return roots[bucketId].root;
    }
}