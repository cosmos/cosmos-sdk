#include <cosmwasm.hpp>
#include <vector>
#include <string>
#include <limits>

// Token denomination constants
namespace denom {
    const std::string ATOM = "uatom";            // Cosmos Hub token
    const std::string IBC_BTC = "ibc/BTC_HASH"; // IBC Bitcoin token
    const std::string IBC_ETH = "ibc/ETH_HASH"; // IBC Ethereum token
    const std::string OSMO = "ibc/OSMO_HASH";   // IBC Osmosis token
    const std::string JUNO = "ibc/JUNO_HASH";   // IBC Juno token
    const std::string STARS = "ibc/STARS_HASH"; // IBC Stargaze token
}

// Recipient structure
struct Recipient {
    std::string address;  // Recipient's wallet address
    uint64_t amount;      // Amount to send (in microATOM or token denomination)

    bool is_valid() const {
        return !address.empty() && amount > 0;
    }
};

// Input message structure
struct ReimburseMsg {
    std::vector<Recipient> recipients; // List of recipients with amounts

    bool is_valid() const {
        if (recipients.empty()) {
            return false;
        }
        for (const auto& recipient : recipients) {
            if (!recipient.is_valid()) {
                return false;
            }
        }
        return true;
    }
};

// Main contract logic
class Contract : public cosmwasm::Contract<Contract> {
private:
    // Helper to calculate the total amount required, with overflow protection
    uint64_t calculate_total_amount(const std::vector<Recipient>& recipients) const {
        uint64_t total = 0;
        for (const auto& recipient : recipients) {
            if (recipient.amount > std::numeric_limits<uint64_t>::max() - total) {
                throw std::overflow_error("Total amount exceeds maximum uint64_t value");
            }
            total += recipient.amount;
        }
        return total;
    }

    // Helper to create bank messages for recipients
    std::vector<cosmwasm::BankMsg> create_bank_msgs(
        const std::vector<Recipient>& recipients,
        const std::string& token_denom
    ) const {
        std::vector<cosmwasm::BankMsg> msgs;
        msgs.reserve(recipients.size()); // Optimize vector growth

        for (const auto& recipient : recipients) {
            msgs.emplace_back(cosmwasm::BankMsg::Send{
                recipient.address,
                {{token_denom, recipient.amount}}
            });
        }
        return msgs;
    }

public:
    cosmwasm::Response execute(const cosmwasm::MessageInfo& info, const ReimburseMsg& msg) {
        // Validate provided funds
        if (info.funds.empty()) {
            throw std::runtime_error("No funds provided");
        }

        const auto& token_denom = info.funds.at(0).denom;

        // Validate token denomination (defaults to ATOM for now)
        if (token_denom != denom::ATOM) {
            throw std::runtime_error("Invalid token denomination: expected " + denom::ATOM);
        }

        // Validate message structure
        if (!msg.is_valid()) {
            throw std::runtime_error("Invalid recipients configuration");
        }

        uint64_t total_required = calculate_total_amount(msg.recipients);
        uint64_t provided_funds = info.funds.at(0).amount;

        if (provided_funds < total_required) {
            throw std::runtime_error("Insufficient funds provided: required " + 
                std::to_string(total_required) + ", got " + std::to_string(provided_funds));
        }

        // Generate BankMsg::Send messages for recipients
        auto bank_msgs = create_bank_msgs(msg.recipients, token_denom);

        // Create response with messages, attributes, and events
        return cosmwasm::Response()
            .add_messages(bank_msgs)
            .add_attribute("action", "reimburse")
            .add_attribute("sender", info.sender)
            .add_attribute("total_amount", std::to_string(total_required))
            .add_attribute("recipient_count", std::to_string(msg.recipients.size()))
            .add_event("reimburse", {
                {"sender", info.sender},
                {"total_amount", std::to_string(total_required)},
                {"denom", token_denom},
                {"recipients", std::to_string(msg.recipients.size())}
            });
    }
};
