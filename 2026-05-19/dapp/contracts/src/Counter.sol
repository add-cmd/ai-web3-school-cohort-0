// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

contract Counter {
    uint256 private _count;
    address public owner;

    event CountChanged(uint256 newCount, address triggeredBy);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);

    constructor() {
        owner = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "Caller is not owner");
        _;
    }

    function count() external view returns (uint256) {
        return _count;
    }

    function increment() external {
        _count++;
        emit CountChanged(_count, msg.sender);
    }

    function decrement() external {
        require(_count > 0, "Count cannot be negative");
        _count--;
        emit CountChanged(_count, msg.sender);
    }

    function setCount(uint256 newCount) external onlyOwner {
        _count = newCount;
        emit CountChanged(_count, msg.sender);
    }

    function transferOwnership(address newOwner) external onlyOwner {
        require(newOwner != address(0), "New owner is zero address");
        emit OwnershipTransferred(owner, newOwner);
        owner = newOwner;
    }
}
