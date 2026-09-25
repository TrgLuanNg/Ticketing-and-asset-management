using System;
using System.Collections.Generic;

namespace SolidOrderService
{
    // ==========================================
    // 1. DOMAIN (Yêu cầu 1)
    // ==========================================
    public enum OrderStatus
    {
        PENDING, PAID, COMPLETED
    }

    public class Order
    {
        public string OrderId { get; private set; }
        public double TotalAmount { get; private set; }
        public OrderStatus Status { get; private set; }
        public string ReferrerId { get; private set; }

        public Order(string orderId, double totalAmount, OrderStatus status, string referrerId)
        {
            OrderId = orderId;
            TotalAmount = totalAmount;
            Status = status;
            ReferrerId = referrerId;
        }

        // Logic nghiệp vụ cốt lõi: Domain từ chối hoàn tất nếu chưa thanh toán
        public void Complete()
        {
            if (Status != OrderStatus.PAID)
            {
                throw new InvalidOperationException("Lỗi: Không thể hoàn tất đơn hàng chưa thanh toán (Chỉ chấp nhận trạng thái PAID).");
            }
            Status = OrderStatus.COMPLETED;
        }

        public override string ToString()
        {
            return $"Order{{id='{OrderId}', amount={TotalAmount}, status={Status}}}";
        }
    }

    // ==========================================
    // 2. INTERFACES (Yêu cầu 2, 3) - Abstractions
    // ==========================================
    public interface IOrderRepository
    {
        void Save(Order order);
    }

    public interface IRewardPolicy
    {
        void ApplyReward(Order order);
    }

    public interface INotificationService
    {
        void SendNotification(Order order);
    }

    // ==========================================
    // 3. IMPLEMENTATIONS
    // ==========================================
    public class MySqlOrderRepository : IOrderRepository
    {
        public void Save(Order order)
        {
            Console.WriteLine($"[Database] Đã lưu cập nhật đơn hàng vào CSDL: {order}");
        }
    }

    public class ReferrerRewardPolicy : IRewardPolicy
    {
        public void ApplyReward(Order order)
        {
            if (!string.IsNullOrEmpty(order.ReferrerId))
            {
                double reward = order.TotalAmount * 0.02; // 2% tổng tiền
                // Sử dụng format {0:N0} để in ra số tiền có dấu phân cách hàng nghìn cho đẹp
                Console.WriteLine($"[Reward] Đã cộng phần thưởng cho người giới thiệu ({order.ReferrerId}): {reward:N0} đồng");
            }
        }
    }

    // Minh họa Yêu cầu 4: Dễ dàng thêm class này mà KHÔNG cần sửa OrderService
    public class VipCustomerRewardPolicy : IRewardPolicy
    {
        public void ApplyReward(Order order)
        {
            Console.WriteLine("[Reward] Đã cộng điểm thưởng tích luỹ cho khách hàng VIP.");
        }
    }

    public class EmailNotificationService : INotificationService
    {
        public void SendNotification(Order order)
        {
            Console.WriteLine("[Notification] Gửi Email: Đơn hàng của bạn đã được hoàn tất thành công!");
        }
    }

    // ==========================================
    // 4. CORE SERVICE (Điều phối chính)
    // ==========================================
    public class OrderService
    {
        private readonly IOrderRepository _repository;
        private readonly IEnumerable<IRewardPolicy> _rewardPolicies;
        private readonly INotificationService _notificationService;

        // Dependency Injection
        public OrderService(IOrderRepository repository,
                            IEnumerable<IRewardPolicy> rewardPolicies,
                            INotificationService notificationService)
        {
            _repository = repository;
            _rewardPolicies = rewardPolicies;
            _notificationService = notificationService;
        }

        public void CompleteOrder(Order order)
        {
            // 1. Cập nhật trạng thái (Ủy quyền cho Domain)
            order.Complete();

            // 2. Lưu đơn hàng
            _repository.Save(order);

            // 3. Xử lý các chính sách phần thưởng (Open/Closed Principle)
            foreach (var policy in _rewardPolicies)
            {
                policy.ApplyReward(order);
            }

            // 4. Gửi thông báo
            _notificationService.SendNotification(order);
        }
    }

    // ==========================================
    // 5. CLIENT / MAIN (Yêu cầu 5)
    // ==========================================
    class Program
    {
        static void Main(string[] args)
        {
            // Cấu hình Dependencies
            IOrderRepository repo = new MySqlOrderRepository();
            INotificationService notifier = new EmailNotificationService();

            // Cấu hình các chính sách thưởng (Dễ dàng mở rộng)
            var policies = new List<IRewardPolicy>
            {
                new ReferrerRewardPolicy()
                // , new VipCustomerRewardPolicy() // Bỏ comment để chạy thêm policy VIP
            };

            OrderService orderService = new OrderService(repo, policies, notifier);

            // Khởi tạo đơn hàng theo đề bài: 1.500.000 đồng, trạng thái PAID
            Order order = new Order("ORD-2023-001", 1500000, OrderStatus.PAID, "REF-USER-09");

            Console.WriteLine("--- TRƯỚC KHI HOÀN TẤT ---");
            Console.WriteLine($"Trạng thái đơn hàng: {order.Status}");
            Console.WriteLine("\n--- BẮT ĐẦU XỬ LÝ ---");

            // Thực thi thuật toán điều phối chính
            orderService.CompleteOrder(order);

            Console.WriteLine("\n--- SAU KHI HOÀN TẤT ---");
            Console.WriteLine($"Trạng thái đơn hàng hiện tại: {order.Status}");

            Console.ReadLine();
        }
    }
}
