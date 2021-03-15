# AnchorPlatform
Factom requires a flexible means of anchoring the Factom Protocol to other ledgers.
The Anchor Platform is designed to meet these needs.  

In the original design of Factom, every block is anchored to the Bitcoin blockchain.  
We also used an "Anchor Master" to control a particular Bitcoin address for writing these
Anchors.  Since only one address is used, a user looking at Bitcoin could in theory
enumerate all the anchor transactions from Bitcoin only and validate the Factom 
Blockchain.  Factom could not fork into another chain without the cooperation of the
Anchor Master, from a Bitcoin perspective.

We added Ethereum as another ledger to which we add anchors. That brings Factom up to 
having two "Anchor Ledgers."  We anticipate adding more Anchor Ledgers to Factom, 
particularly as fees in Ethereum and Bitcoin rise.

This is an advantage, but we would like the anchoring done for Bitcoin to be distributed
to multiple parties.  In the design of the Anchor Platform, we will make all the 
ANOs potential "Anchor Masters" able to write anchors to the various Anchor ledgers.

Additionally, the need to manage anchors is significant. Providing users with clean, 
fast and efficient anchors is critical to Factom's value proposition. The Anchor
Platform will provide a UI to view Anchoring against all Anchor Ledgers as they
are written.  Anchor Masters will be able to use the Anchor Platform to manage
the Anchors written.

For users of the Factom Protocol. the Anchor platform will provide UI and APIs
to generate Anchor proofs against Anchor Ledgers and to later validate those proofs.
