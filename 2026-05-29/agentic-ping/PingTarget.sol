// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract PingTarget {
    uint256 public pingCount;
    
    // 记录是谁（哪个 Agent）在什么时间执行了动作
    event Pinged(address indexed agent, uint256 timestamp);

    function ping() external {
        pingCount++;
        emit Pinged(msg.sender, block.timestamp);
    }
}