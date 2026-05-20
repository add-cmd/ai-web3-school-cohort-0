// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Script} from "forge-std/Script.sol";
import {TokenShop} from "../src/TokenShop.sol";

contract DeployTokenShop is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        vm.startBroadcast(deployerPrivateKey);
        TokenShop shop = new TokenShop();
        // 设置代币价格 — 允许 SimpleToken 购买
        shop.setPrice(0x62E3395eCFa2d18afB8F0cfbB1FA55948Dd03674, 1 ether);
        vm.stopBroadcast();
    }
}
