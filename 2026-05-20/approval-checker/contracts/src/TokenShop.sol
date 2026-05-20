// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

/// @title TokenShop — 模拟一个需要 approve 的 dApp（DEX / NFT 市场）
/// @dev 用户需要先 approve，才能调用 buyItem 或 deposit
contract TokenShop {
    address public owner;
    mapping(address => uint256) public prices;       // token -> price in wei
    mapping(uint256 => address) public itemSeller;   // itemId -> seller

    event ItemPurchased(uint256 indexed itemId, address buyer, uint256 price);
    event PriceSet(address indexed token, uint256 price);

    constructor() {
        owner = msg.sender;
    }

    /// 设置某种代币的购买价格（模拟 dApp 功能）
    function setPrice(address token, uint256 price) external {
        prices[token] = price;
        emit PriceSet(token, price);
    }

    /// 用户 approve 后调用此方法购买
    function buyItem(address token, uint256 amount) external returns (bool) {
        require(prices[token] > 0, "token not supported");
        // dApp 会 transferFrom 用户
        IERC20(token).transferFrom(msg.sender, owner, amount);
        return true;
    }
}

interface IERC20 {
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
}
