// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Script} from "forge-std/Script.sol";
import {SimpleToken} from "../src/SimpleToken.sol";

contract DeploySimpleToken is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        vm.startBroadcast(deployerPrivateKey);
        new SimpleToken(1_000_000 ether);
        vm.stopBroadcast();
    }
}
