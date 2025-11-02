"""Payment tools for the LangGraph agent.

Exposes MiniPay backend operations as LangGraph-compatible tools with proper
error handling and response formatting.
"""

from __future__ import annotations

from typing import Any, Callable, List, Optional, cast
import json
import uuid

from langchain_core.tools import tool
from langchain_core.language_model import BaseLLM

from react_agent.minipay_client import MiniPayClient


@tool
async def create_charge(
    amount: int,
    currency: str = "usd",
    customer: str = "",
    idempotency_key: Optional[str] = None,
) -> str:
    """Create a new payment charge in the MiniPay system.

    Args:
        amount: Amount in cents (e.g., 1000 = $10.00)
        currency: Currency code (default: usd)
        customer: Customer identifier (required for tracking)
        idempotency_key: Optional idempotency key to prevent duplicate charges.
                        If not provided, one will be generated.

    Returns:
        JSON string containing transaction details including ID, status, and creation time.

    Example:
        create_charge(amount=5000, currency="usd", customer="cust_123")
        Returns: {"id": "txn_...", "amount": 5000, "status": "succeeded", ...}
    """
    try:
        if not idempotency_key:
            idempotency_key = f"agent_{uuid.uuid4().hex[:12]}"

        async with MiniPayClient() as client:
            response = await client.create_charge(
                amount=amount,
                currency=currency,
                customer=customer,
                idempotency_key=idempotency_key,
            )
            return json.dumps(
                {
                    "success": True,
                    "transaction_id": response.id,
                    "amount": response.amount,
                    "currency": response.currency,
                    "customer": response.customer,
                    "status": response.status,
                    "idempotency_key": response.idempotency_key,
                    "created_at": response.created_at,
                }
            )
    except Exception as e:
        return json.dumps({"success": False, "error": str(e)})


@tool
async def refund_charge(transaction_id: str) -> str:
    """Refund a previously created charge.

    Args:
        transaction_id: The transaction ID to refund (starts with 'txn_')

    Returns:
        JSON string containing refund details including status and refund timestamp.

    Example:
        refund_charge("txn_550e8400e29b41d4a716446655440000")
        Returns: {"id": "txn_...", "status": "refunded", "refunded_at": "...", ...}
    """
    try:
        async with MiniPayClient() as client:
            response = await client.refund_transaction(transaction_id)
            return json.dumps(
                {
                    "success": True,
                    "transaction_id": response.id,
                    "amount": response.amount,
                    "status": response.status,
                    "refunded_at": response.refunded_at,
                }
            )
    except Exception as e:
        return json.dumps({"success": False, "error": str(e)})


@tool
async def get_account_balance() -> str:
    """Get the current account balance.

    Calculates balance as: (successful charges) - (refunded amounts)

    Returns:
        JSON string containing successful transactions, refunded transactions, and total balance.

    Example:
        get_account_balance()
        Returns: {"balance": 50000, "successful_transactions": 10, "refunded_transactions": 2}
    """
    try:
        async with MiniPayClient() as client:
            response = await client.get_balance()
            return json.dumps(
                {
                    "success": True,
                    "balance": response.balance,
                    "successful_transactions": response.successful_transactions,
                    "refunded_transactions": response.refunded_transactions,
                }
            )
    except Exception as e:
        return json.dumps({"success": False, "error": str(e)})


@tool
async def get_system_metrics() -> str:
    """Get system metrics for the MiniPay backend.

    Returns statistics about charges, refunds, and webhook delivery.

    Returns:
        JSON string containing metrics for charges, refunds, webhooks, and retries.

    Example:
        get_system_metrics()
        Returns: {
            "total_charges": 100,
            "total_refunds": 5,
            "pending_webhooks": 3,
            "delivered_webhooks": 95,
            "webhook_retries": 12
        }
    """
    try:
        async with MiniPayClient() as client:
            response = await client.get_metrics()
            return json.dumps(
                {
                    "success": True,
                    "total_charges": response.total_charges,
                    "total_refunds": response.total_refunds,
                    "pending_webhooks": response.pending_webhooks,
                    "delivered_webhooks": response.delivered_webhooks,
                    "failed_webhooks": response.failed_webhooks,
                    "webhook_retries": response.webhook_retries,
                }
            )
    except Exception as e:
        return json.dumps({"success": False, "error": str(e)})


@tool
async def check_backend_health() -> str:
    """Check the health status of the MiniPay backend.

    Returns:
        JSON string indicating if the backend is healthy and operational.

    Example:
        check_backend_health()
        Returns: {"healthy": true, "status": "The MiniPay backend is operational"}
    """
    try:
        async with MiniPayClient() as client:
            is_healthy = await client.health_check()
            status = "The MiniPay backend is operational" if is_healthy else "The MiniPay backend is unavailable"
            return json.dumps(
                {
                    "success": True,
                    "healthy": is_healthy,
                    "status": status,
                }
            )
    except Exception as e:
        return json.dumps({"success": False, "error": str(e)})


PAYMENT_TOOLS: List[Callable[..., Any]] = [
    create_charge,
    refund_charge,
    get_account_balance,
    get_system_metrics,
    check_backend_health,
]
